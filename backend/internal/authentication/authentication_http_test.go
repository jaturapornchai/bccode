package authentication

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"smlcloudplatform/internal/config"
	"strconv"
	"strings"
	"testing"
)

func TestAuthenticationRouteSurface(t *testing.T) {
	routes := registeredAuthenticationRoutes(t)

	for _, route := range []string{
		"POST /login",
		"POST /googlelogin",
		"POST /refresh",
		"POST /logout",
		"POST /dev-login",
	} {
		if !routes[route] {
			t.Errorf("required authentication route is not registered: %s", route)
		}
	}

	for _, route := range []string{
		"POST /poslogin",
		"POST /login/email",
		"POST /login/phone-number",
		"POST /login/line",
		"POST /linelogin",
		"POST /tokenlogin",
		"POST /register",
		"POST /send-phonenumber-otp",
		"POST /forgot-password-phonenumber",
		"POST /register-phonenumber",
		"POST /register/exists-phonenumber",
		"POST /register/exists-username",
		"PUT /profile/password/reset/:username",
	} {
		if routes[route] {
			t.Errorf("disabled authentication route is still registered: %s", route)
		}
	}
}

func TestDevLoginConfigGate(t *testing.T) {
	validSecret := strings.Repeat("s", 32)
	tests := []struct {
		name                  string
		dataEnvironment       string
		environmentConfigured bool
		enabled               string
		userUID               string
		secret                string
		wantEnabled           bool
	}{
		{name: "explicit dev with complete config", dataEnvironment: config.DataEnvironmentDev, environmentConfigured: true, enabled: "true", userUID: " dev-user ", secret: validSecret, wantEnabled: true},
		{name: "unknown environment must fail closed", dataEnvironment: "development", environmentConfigured: true, enabled: "true", userUID: "dev-user", secret: validSecret},
		{name: "environment absent", dataEnvironment: config.DataEnvironmentDev, environmentConfigured: false, enabled: "true", userUID: "dev-user", secret: validSecret},
		{name: "production", dataEnvironment: config.DataEnvironmentPRO, environmentConfigured: true, enabled: "true", userUID: "dev-user", secret: validSecret},
		{name: "feature disabled", dataEnvironment: config.DataEnvironmentDev, environmentConfigured: true, enabled: "false", userUID: "dev-user", secret: validSecret},
		{name: "user uid absent", dataEnvironment: config.DataEnvironmentDev, environmentConfigured: true, enabled: "true", secret: validSecret},
		{name: "secret too short", dataEnvironment: config.DataEnvironmentDev, environmentConfigured: true, enabled: "true", userUID: "dev-user", secret: strings.Repeat("s", 31)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, enabled := devLoginConfigFor(tt.dataEnvironment, tt.environmentConfigured, tt.enabled, tt.userUID, tt.secret)
			if enabled != tt.wantEnabled {
				t.Fatalf("enabled = %v, want %v", enabled, tt.wantEnabled)
			}
			if tt.wantEnabled && (got.userUID != "dev-user" || got.secret != validSecret) {
				t.Fatalf("unexpected enabled config: %#v", got)
			}
		})
	}
}

func TestDevLoginSecretComparison(t *testing.T) {
	secret := strings.Repeat("s", 32)
	if !devLoginSecretMatches(secret, secret) {
		t.Fatal("matching secret must pass")
	}
	if devLoginSecretMatches(secret, strings.Repeat("x", 32)) {
		t.Fatal("different secret must fail")
	}
	if devLoginSecretMatches(secret, secret+"x") {
		t.Fatal("different-length secret must fail")
	}
}

func registeredAuthenticationRoutes(t *testing.T) map[string]bool {
	t.Helper()
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate authentication route test")
	}
	sourceFile := filepath.Join(filepath.Dir(testFile), "authentication_http.go")
	parsed, err := parser.ParseFile(token.NewFileSet(), sourceFile, nil, 0)
	if err != nil {
		t.Fatalf("parse authentication routes: %v", err)
	}

	routes := map[string]bool{}
	ast.Inspect(parsed, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		switch selector.Sel.Name {
		case "GET", "POST", "PUT", "PATCH", "DELETE":
		default:
			return true
		}
		pathLiteral, ok := call.Args[0].(*ast.BasicLit)
		if !ok || pathLiteral.Kind != token.STRING {
			return true
		}
		path, err := strconv.Unquote(pathLiteral.Value)
		if err == nil {
			routes[selector.Sel.Name+" "+path] = true
		}
		return true
	})
	return routes
}

func TestSelectableCompanyFilterRequiresActiveCompany(t *testing.T) {
	filter := selectableCompanyFilter("holdingtest", "COMP-A")

	if filter["holdingcode"] != "holdingtest" {
		t.Fatalf("holdingcode filter mismatch: %v", filter["holdingcode"])
	}
	if filter["code"] != "COMP-A" {
		t.Fatalf("company code filter mismatch: %v", filter["code"])
	}
	if filter["isactive"] != true {
		t.Fatalf("inactive companies must not be selectable: %v", filter["isactive"])
	}
	if _, ok := filter["deletedat"]; !ok {
		t.Fatal("soft-deleted companies must not be selectable")
	}
}
