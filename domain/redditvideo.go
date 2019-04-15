package domain

import (
	"net/url"
	"os"
	"strings"

	"github.com/peter-jozsa/jsonpath"
	"gopkg.in/mgo.v2/bson"

	"github.com/johnwyles/vrddt-droplets/pkg/errors"
)

// TODO: Do we add AudioURL, Title, and VideoURL, and VrddtVideo to Validate()?

// TODO: Incorporate errors into pkg/errors

const (
	// JSONPathForTitle is the JSON path to find the title for a Reddit post
	JSONPathForTitle = `$.data.children[0].data.title`

	// JSONPathForVideoURL is the JSON path to find the video URL for a Reddit post
	JSONPathForVideoURL = `$.data.children[0].data.media.reddit_video.fallback_url`
)

var (
	// ErrJSONTitle is the error returned when the JSON does not parse in order to find the title
	ErrJSONTitle = errors.New("JSON data does not have exactly one match for the Title: " + JSONPathForTitle)

	// ErrJSONVideoURL is the error returned when the JSON does not parse in order to find the video URL
	ErrJSONVideoURL = errors.New("JSON data does not have exactly one match for the video URL: " + JSONPathForVideoURL)

	// ErrNotDASH is the error returned when the video URL found when
	// attempting to set the audio URL is not a URL containing "DASH_"
	ErrNotDASH = errors.New("The Reddit video URL does not seem to contain a DASH video")

	// KnownRedditDomains are all of the known Reddit domains prefixed with a "." so
	// that we do not process any requests for domains which contain a known
	// Reddit domain name at the end of them (e.g. "foo-reddit.com")
	KnownRedditDomains = []string{
		".redd.it",
		".reddit.com",
		".redditstatic.com",
	}

	// RedditDomainURLPrefix will be what to prepend to arbitrary URIs that
	// come in by a request that we will attempt to locate a valid Reddit URL
	RedditDomainURLPrefix = "https://reddit.com"

	// TemporaryAudioFilePrefix is the file prefix for the file that will hold
	// the contents of the audio downloaded
	TemporaryAudioFilePrefix = "vrddt-input-reddit-audio*.mp4"

	// TemporaryVideoFilePrefix is the file prefix for the file that will hold
	// the contents of the video downloaded
	TemporaryVideoFilePrefix = "vrddt-input-reddit-video*.mp4"
)

// RedditAudio is nothing more than a simpl struct for a file which we may or
// may not consider any attributes of important in the future, we use this
// since we are merging both Audio and Video for this project but who knows
// what we may want to do with it in the future
type RedditAudio struct {
	FilePath   string   `json:"-" bson:"-"`
	FileHandle *os.File `json:"-" bson:"-"`
}

// RedditVideo represents information about registered reddit videos.
type RedditVideo struct {
	// AudioURL is the URL to the audio for the reddit video.
	AudioURL string `json:"audio_url,omitempty" bson:"audio_url,omitempty"`

	RedditAudio *RedditAudio `json:"-" bson:"-"`

	FilePath string `json:"-" bson:"-"`

	FileHandle *os.File `json:"-" bson:"-"`

	// Meta holds the generic information about the vrddt video.
	Meta `json:",inline,omitempty" bson:",inline,omitempty"`

	// URL should contain a valid URL for the reddit link for the reddit video.
	URL string `json:"url,omitempty" bson:"url,omitempty"`

	// Title is the title of the reddit link for the reddit video.
	Title string `json:"title,omitempty" bson:"title,omitempty"`

	// VideoURL is the URL to the video for the reddit video.
	VideoURL string `json:"video_url,omitempty" bson:"video_url,omitempty"`

	// VrddtVideo represents the ID of the vrddt video that we created with this video.
	VrddtVideoID bson.ObjectId `json:"vrddt_video_id,omitempty" bson:"vrddt_video_id,omitempty"`
}

// NewRedditVideo will return a new Reddit video.
func NewRedditVideo() *RedditVideo {
	return &RedditVideo{
		Meta: NewMeta(),
	}
}

// Download will perform DownloadVideo() and DownloadAudio() ignoring an error
// in DownloadAudio if present
func (r *RedditVideo) Download() (err error) {
	if err = r.DownloadVideo(); err != nil {
		return
	}

	_ = r.DownloadAudio()

	return
}

// DownloadVideo will download the video file to a particular path
func (r *RedditVideo) DownloadVideo() (err error) {
	r.FileHandle, err = DownloadToTemporaryFile(r.VideoURL, TemporaryVideoFilePrefix)
	r.FilePath = r.FileHandle.Name()

	return
}

// DownloadAudio will download the audio file to a particular path
func (r *RedditVideo) DownloadAudio() (err error) {
	r.RedditAudio = &RedditAudio{}
	r.RedditAudio.FileHandle, err = DownloadToTemporaryFile(r.AudioURL, TemporaryAudioFilePrefix)
	if err != nil {
		// r.log.Debug().Err(err).Msg("reddit video does not have associated audio or there was an issue retrieving it")
		r.RedditAudio = nil
		r.AudioURL = ""
	} else {
		r.RedditAudio.FilePath = r.RedditAudio.FileHandle.Name()
	}

	return
}

// SetAudioURL returns the URL to the audio for a given Reddit URL
func (r *RedditVideo) SetAudioURL() (err error) {
	if strings.Contains(r.VideoURL, "DASH_") {
		r.AudioURL = strings.Split(r.VideoURL, "DASH_")[0] + "audio"
	} else {
		err = ErrNotDASH
	}

	return
}

// SetFinalURL will set the URL as the final URL after less than or equal to
// util.RedirectMax http redirects for the supplied URL
func (r *RedditVideo) SetFinalURL() (err error) {
	r.URL, err = GetFinalURL(r.URL)

	return
}

// SetMetadata sets the title, video URL and audio URL for a given Reddit video from
// a Reddit URL
func (r *RedditVideo) SetMetadata() (err error) {
	jsonURL := r.getJSONURL()

	jsonData, err := GetJSONData(jsonURL)
	if err != nil {
		return
	}

	r.Title, err = getTitleFromJSONData(jsonData)
	if err != nil {
		return
	}

	r.VideoURL, err = getVideoURLFromJSONData(jsonData)
	if err != nil {
		return
	}

	err = r.SetAudioURL()

	return
}

// Validate performs basic validation of user information.
func (r RedditVideo) Validate() error {
	// _, err := url.ParseRequestURI(redditVideo.AudioURL)
	// if err != nil {
	// 	return errors.InvalidValue("AudioURL", err.Error())
	// }

	// if err := r.Meta.Validate(); err != nil {
	// 	return err
	// }

	// if len(strings.TrimSpace(string(redditVideo.Title))) == 0 {
	// 	return errors.MissingField("Title")
	// }

	_, err := url.ParseRequestURI(r.URL)
	if err != nil {
		return errors.InvalidValue("URL", err.Error())
	}

	// _, err = url.ParseRequestURI(redditVideo.VideoURL)
	// if err != nil {
	// 	return errors.InvalidValue("VideoURL", err.Error())
	// }

	// if !redditVideo.VrddtVideo.Valid() {
	// 	return errors.InvalidValue("VrddtVideo", "Not a valid bson.ObjectId")
	// }

	return nil
}

// getJSONURL will return the JSON URL for a Reddit URL
func (r *RedditVideo) getJSONURL() (redditURL string) {
	redditURL = strings.TrimRight(r.URL, "/") + ".json"

	return
}

// setTitle returns the Title of the given Reddit URL
func (r *RedditVideo) setTitle() (err error) {
	jsonURL := r.getJSONURL()
	jsonData, err := GetJSONData(jsonURL)
	if err != nil {
		return
	}

	r.Title, err = getTitleFromJSONData(jsonData)

	return
}

// setVideoURL returns the URL to the video for a given Reddit URL
func (r *RedditVideo) setVideoURL() (err error) {
	jsonURL := r.getJSONURL()
	jsonData, err := GetJSONData(jsonURL)
	if err != nil {
		return
	}

	r.VideoURL, err = getVideoURLFromJSONData(jsonData)

	return
}

// getTitleFromJSONData will hunt down the title for a Reddit post from a
// JSON path
func getTitleFromJSONData(jsonData interface{}) (title string, err error) {
	pattern, _ := jsonpath.Compile(JSONPathForTitle)
	patternMatches, err := pattern.Lookup(jsonData)
	if err != nil {
		return
	}

	matches := patternMatches.([]interface{})
	if len(matches) != 1 {
		return title, ErrJSONTitle
	}

	title = matches[0].(string)

	return
}

// getVideoURLFromJSONData will hunt down the video URL for a Reddit post from
// a JSON path
func getVideoURLFromJSONData(jsonData interface{}) (videoURL string, err error) {
	pattern, _ := jsonpath.Compile(JSONPathForVideoURL)
	patternMatches, err := pattern.Lookup(jsonData)
	if err != nil {
		return
	}

	matches := patternMatches.([]interface{})
	if len(matches) != 1 {
		return videoURL, ErrJSONVideoURL
	}

	videoURL = matches[0].(string)

	return
}

// isRedditURL will validate that whatever has been thrown at us is actually a
// Reddit URL and saves us time and trouble
func isRedditURL(originalURL string) (valid bool) {
	valid = false
	u, err := url.Parse(originalURL)
	if err != nil {
		return
	}

	host := strings.ToLower(u.Host)
	for _, validDomain := range KnownRedditDomains {
		if strings.HasSuffix(host, validDomain) {
			return true
		} else if host == strings.TrimPrefix(validDomain, ".") {
			return true
		}
	}

	return
}
