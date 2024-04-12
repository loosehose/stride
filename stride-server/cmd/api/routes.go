package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	// Create a router mux
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(app.enableCORS)

	// Existing routes
	mux.Get("/", app.Home)
	mux.Get("/agents", app.AllAgents)
	mux.Get("/domains", app.AllDomains)

	// Route for creating an agent
	mux.Post("/agents", app.createAgentHandler)

	// New route for deleting an agent
	mux.Delete("/agents/{dropletID}", app.deleteAgentHandler)

	// Route for deleting a subdomain
	mux.Delete("/subdomains/{domainName}/{subdomainName}", app.deleteSubdomainHandler)

	// Route for listing all SSH keys
	mux.Get("/ssh-keys", app.listSSHKeysHandler)

	// Route for creating a subdomain
	mux.Post("/subdomains", app.createSubdomainHandler)

	// New route for setting up Teamserver
	mux.Post("/teamserver-setup", app.setupTeamserverHandler)

	// New route for setting up Redirector
	mux.Post("/redirector-setup", app.setupRedirectorHandler)

	// New route for setting up port forwarding
	mux.Post("/port-forwarding-setup", app.setupPortForwardingHandler)

	mux.Post("/phishing-setup", app.setupPhishingHandler)

	// Route for Payloads
	mux.Post("/payload-setup", app.setupPayloadHandler)

	// WebSocket route
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		app.wsManager.Handler(w, r)
	})

	return mux
}
