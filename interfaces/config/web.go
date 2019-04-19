package config

// WebConfig stores the configuration for the web server
type WebConfig struct {
	Address         string
	CertFile        string
	GracefulTimeout int
	KeyFile         string
	StaticDir       string
	TemplateDir     string
	VrddtAPIURI     string
}
