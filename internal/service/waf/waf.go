package service_waf

import (
	"log"
	"os"
	"strings"
	"sync"

	"github.com/corazawaf/libinjection-go"
	"github.com/jahrulnr/go-waf/config"
	"github.com/jahrulnr/go-waf/internal/interface/service"
	"github.com/jahrulnr/go-waf/pkg/logger"
	"gopkg.in/yaml.v2"
)

type WAFService struct {
	config *config.Config

	commandInjectionKeywords []string
	pathTraversalKeywords    []string
}

func NewWAFService(config *config.Config, keywordsFile string) service.WAFInterface {
	keywords := loadKeywords(keywordsFile)

	return &WAFService{
		config: config,

		commandInjectionKeywords: keywords.CommandInjectionKeywords,
		pathTraversalKeywords:    keywords.PathTraversalKeywords,
	}
}

func loadKeywords(filename string) service.Keywords {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("error reading keywords file: %v", err)
	}

	var keywords service.Keywords
	err = yaml.Unmarshal(data, &keywords)
	if err != nil {
		log.Fatalf("error unmarshalling keywords: %v", err)
	}

	return keywords
}

func (w *WAFService) HandleRequest(request *service.Request) (*service.Response, error) {
	var headerThreatDetected, bodyThreatDetected bool
	var wg sync.WaitGroup

	// Check for header threats
	if w.config.WAF_PROTECT_HEADER {
		wg.Add(1)
		go func() {
			defer wg.Done()
			headerThreatDetected = w.DetectHeaderThreats(request)
		}()
	}

	// Check for body threats
	if w.config.WAF_PROTECT_BODY {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bodyThreatDetected = w.DetectBodyThreats(request)
		}()
	}

	// Wait for all checks to complete
	wg.Wait()

	// If any threats are detected, return a 403 response
	if headerThreatDetected || bodyThreatDetected {
		return &service.Response{StatusCode: 403, Body: []byte("Threat Detected")}, nil
	}

	// Return a successful response if no threats are detected
	return nil, nil
}

func (w *WAFService) DetectHeaderThreats(request *service.Request) bool {
	// Check for SQL injection patterns in headers
	for _, value := range request.Headers {
		if injection, _ := libinjection.IsSQLi(value); injection {
			return true
		}
	}

	// Check for XSS patterns in headers
	for _, value := range request.Headers {
		if libinjection.IsXSS(value) {
			return true
		}
	}

	// Check for command injection patterns in headers
	for _, pattern := range w.commandInjectionKeywords {
		for _, value := range request.Headers {
			if strings.Contains(value, pattern) {
				logger.Logger("Threat Detected", pattern).Warn()
				return true
			}
		}
	}

	// Check for path traversal patterns in headers (if applicable)
	for _, pattern := range w.pathTraversalKeywords {
		for _, value := range request.Headers {
			if strings.Contains(value, pattern) {
				logger.Logger("Threat Detected", pattern).Warn()
				return true
			}
		}
	}

	return false
}

func (w *WAFService) DetectBodyThreats(request *service.Request) bool {
	// Check for SQL injection patterns in headers
	if injection, _ := libinjection.IsSQLi(string(request.Body)); injection {
		logger.Logger("Threat Detected (SQL Injection)", string(request.Body)).Warn()
		return true
	}

	// Check for XSS patterns in headers
	if libinjection.IsXSS(string(request.Body)) {
		logger.Logger("Threat Detected (XSS Attact)", string(request.Body)).Warn()
		return true
	}

	// Check for command injection patterns
	for _, pattern := range w.commandInjectionKeywords {
		if strings.Contains(string(request.Body), pattern) {
			logger.Logger("Threat Detected", pattern).Warn()
			return true
		}
	}

	// Check for path traversal patterns
	for _, pattern := range w.pathTraversalKeywords {
		if strings.Contains(string(request.Body), pattern) {
			logger.Logger("Threat Detected", pattern).Warn()
			return true
		}
	}

	return false
}
