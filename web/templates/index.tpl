<!doctype html>
<html lang="en">

<head>
  <!-- Required meta tags -->
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">

  <!-- Bootstrap CSS -->
  <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.1.3/css/bootstrap.min.css" integrity="sha384-MCw98/SFnGE8fJT3GXwEOngsV7Zt27NXFoaoApmYm81iuXoPkFOJwJ8ERdknLPMO"
    crossorigin="anonymous">

  <title>vrddt Web Server</title>
</head>

<body>
  <div class="container">
    <nav class="navbar navbar-light bg-light">
       <a class="navbar-brand">
          vrddt Web Server
       </a>
    </nav>
  </div>
  <div class="container">
    <h1>Enter Reddit URL:</h1>
    <form id="convert-form" class="form-inline">
      <input id="url" name="url" class="form-control mr-sm-2" type="text" placeholder="https://" aria-label="Convert" value="{{.RedditURL}}" />
      <button class="btn btn-outline-success my-2 my-sm-0" type="submit">Convert</button>
    </form>
    <div id="convert-response">
      <h1>Converted Video Link:</h1>
      <a id="convert-link" href="{{.VrddtURL}}">{{.VrddtURL}}</a>
    </div>
  </div>
  <!-- Optional JavaScript -->
  <!-- jQuery first, then Popper.js, then Bootstrap JS -->
  <script src="https://code.jquery.com/jquery-3.4.0.min.js" integrity="sha256-BJeo0qm959uMBGb65z40ejJYGSgR7REI4+CW1fNKwOg="
    crossorigin="anonymous"></script>
  <script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.14.3/umd/popper.min.js" integrity="sha384-ZMP7rVo3mIykV+2+9J3UJ46jBk0WLaUAdn689aCwoqbBJiSnjAK/l8WvCWPIPm49"
    crossorigin="anonymous"></script>
  <script src="https://stackpath.bootstrapcdn.com/bootstrap/4.1.3/js/bootstrap.min.js" integrity="sha384-ChfqqxuZUCnJSK3+MXmPNIyE6ZbWh2IMqE241rYiqJxyMiZ6OW/JmZQ5stwEULTy"
    crossorigin="anonymous"></script>
  <script type="text/javascript">
    $('#convert-form').submit(function(e){
      e.preventDefault();
      var reddit_response = "";
      $.ajax({
        data: $(this).serializeArray(),
        dataType: "json",
        type: "GET",
        url: 'https://{{.VrddtAPIURI}}/reddit_videos/',
        success: function(reddit_response)
        {
          var vrddt_response = "";
          $.ajax({
            type: "GET", 
            url: 'https://{{.VrddtAPIURI}}/vrddt_videos/'+reddit_response.vrddt_video_id,
            success: function(vrddt_response)
            {
              console.log(reddit_response);
              console.log(vrddt_response);
              $('a#convert-link').html(reddit_response.title);
              $('a#convert-link').attr('href', vrddt_response.url);
              $('div#convert-response').attr('style', '');
            }
          });
        }
      });
    });
  </script>
</body>

</html>
