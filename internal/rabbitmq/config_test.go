package rabbitmq

import "testing"

func TestWithDefaults_HostPortEmbedded(t *testing.T) {
	cfg := Config{Host: "localhost:5673"}.withDefaults()

	if cfg.Host != "localhost" {
		t.Fatalf("expected host localhost, got %q", cfg.Host)
	}
	if cfg.Port != 5673 {
		t.Fatalf("expected port 5673, got %d", cfg.Port)
	}
}

func TestWithDefaults_ExplicitPortWins(t *testing.T) {
	cfg := Config{Host: "localhost:5673", Port: 5674}.withDefaults()

	if cfg.Host != "localhost" {
		t.Fatalf("expected host localhost, got %q", cfg.Host)
	}
	if cfg.Port != 5674 {
		t.Fatalf("expected port 5674, got %d", cfg.Port)
	}
}

func TestSplitHostPort(t *testing.T) {
	host, port, ok := splitHostPort("localhost:5672")
	if !ok {
		t.Fatal("expected splitHostPort to parse localhost:5672")
	}
	if host != "localhost" || port != 5672 {
		t.Fatalf("unexpected split result host=%q port=%d", host, port)
	}
}
