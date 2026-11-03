package serve

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/pkg/client"
)

//go:embed templates/*
var templateFS embed.FS

var (
	homeTmpl *template.Template
	tagTmpl  *template.Template
)

var port string

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start a local web server showing release information",
	Long:  `Starts a local HTTP server that serves release information as web pages`,
	RunE:  runServe,
}

func init() {
	ServeCmd.Flags().StringVar(&port, "port", "8080", "Port to listen on")
}

func parsePageTemplate(files ...string) (*template.Template, error) {
	return template.ParseFS(templateFS, files...)
}

func runServe(cmd *cobra.Command, args []string) error {
	var err error
	homeTmpl, err = parsePageTemplate("templates/base.html", "templates/home.html")
	if err != nil {
		return fmt.Errorf("failed to parse home template: %w", err)
	}
	tagTmpl, err = parsePageTemplate("templates/base.html", "templates/tag.html")
	if err != nil {
		return fmt.Errorf("failed to parse tag template: %w", err)
	}

	c, err := client.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	staticFS, err := fs.Sub(templateFS, "templates/static")
	if err != nil {
		return fmt.Errorf("failed to create static file system: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	mux.HandleFunc("GET /{$}", handleHome())
	mux.HandleFunc("GET /tag", handleTag(c))

	addr := net.JoinHostPort("localhost", port)
	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx := cmd.Context()
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("Server listening on http://%s", addr)
	err = server.ListenAndServe()
	if err == http.ErrServerClosed {
		<-shutdownDone
		return nil
	}
	return err
}
