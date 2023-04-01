package main

import (
	config "cf-oidc-proxy/config"
	"cf-oidc-proxy/services"
	"cf-oidc-proxy/types"
	"cf-oidc-proxy/validators"
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/cloudflare/cloudflare-go"
	"github.com/coreos/go-oidc/v3/oidc"
	_ "github.com/zitadel/oidc/v2/pkg/oidc"
	"gopkg.in/square/go-jose.v2/json"
	"os"
	"strings"
)

func GenerateResponse(Body *cloudflare.APIToken, Code int) events.APIGatewayProxyResponse {
	body, err := json.Marshal(Body)
	if err != nil {
		fmt.Printf("error marshalling response: %v\n", err)
		return events.APIGatewayProxyResponse{StatusCode: 400}
	}
	return events.APIGatewayProxyResponse{Body: string(body), StatusCode: Code}
}

func GenerateResponseStatusCode(Code int) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{StatusCode: Code}
}

func HandleRequest(_ context.Context, request events.LambdaFunctionURLRequest) (events.APIGatewayProxyResponse, error) {
	requestBody, err := pareRequestBody(request)
	if err != nil {
		fmt.Printf("error parsing request body: %v\n", err)
		return GenerateResponseStatusCode(400), nil
	}
	roleName := requestBody.RoleToAssume
	authorizationHeader := request.Headers["Authorization"]
	addr := strings.Split(request.Headers["X-Forwarded-For"], ",")

	// Load the configuration
	fmt.Println("Loading configuration")
	cfg, err := config.LoadConfiguration()
	if err != nil {
		fmt.Printf("error loading configuration: %v\n", err)
		return GenerateResponseStatusCode(500), nil
	}

	// Load the role configuration
	fmt.Println("Loading roles configuration")
	rcfg, err := config.LoadRoleConfiguration()
	if err != nil {
		fmt.Printf("error loading role configuration: %v\n", err)
		return GenerateResponseStatusCode(500), nil
	}

	// Validate the authorization header and return the id token
	fmt.Println("Validating authorization header")
	idToken, err := validateAuthorization(authorizationHeader, cfg.OidcProxy.Issuer)
	if err != nil {
		fmt.Printf("error validating authorization: %v\n", err)
		return GenerateResponseStatusCode(401), nil
	}

	// Search role by name in the group configuration
	group, err := rcfg.GetRoleByName(roleName)
	if err != nil {
		return GenerateResponseStatusCode(400), nil
	}

	// Match the claims with configure
	match, err := validators.RoleEntityMatcher(*rcfg, roleName, idToken)
	if match != true || err != nil {
		fmt.Printf("error id token not match with requested group")
		return GenerateResponseStatusCode(401), nil
	}
	fmt.Printf("id token match with requested group: %v\n", match)
	// Create Cloudflare client to interact with the cloudflare API
	cfClient, err := services.NewCloudflareClient(cfg.OidcProxy)
	if err != nil {
		fmt.Printf("error creating cloudflare client: %v\n", err)
		return GenerateResponseStatusCode(500), nil
	}

	// Create a short-lived token with the group and the configured permissions
	token, err := cfClient.CreateShortLivedToken(group, roleName, addr)
	if err != nil {
		fmt.Printf("error creating cloudflare short lived token: %v\n", err)
		return GenerateResponseStatusCode(500), nil
	}

	return GenerateResponse(token, 200), nil
}

// Validate the authorization header against the issuer JWKs public key
// Check if token not expired and return the id token
//
// DANGEROUS: If you set the environment variable SKIP_EXPIRY_CHECK to true, the expiry check will be skipped
// This is useful for testing purposes, but should never be used in production
func validateAuthorization(authorizationHeader string, issuer string) (idToken *oidc.IDToken, err error) {
	if authorizationHeader == "" {
		return nil, errors.New("authorization header is empty")
	}
	if authorizationHeader[0:6] != "Bearer" {
		return nil, errors.New("authorization header does not start with Bearer")
	}

	skipExpiryCheck := os.Getenv("SKIP_EXPIRY_CHECK") == "true"
	if skipExpiryCheck == true {
		fmt.Printf("\n!!!!!! DANGEROUS: Skipping expiry check !!!!!!!\n")
	}

	provider, err := oidc.NewProvider(context.Background(), issuer)
	verifier := provider.Verifier(&oidc.Config{SkipClientIDCheck: true, SkipExpiryCheck: skipExpiryCheck})

	res, err := verifier.Verify(context.Background(), authorizationHeader[7:])
	if err != nil {
		return nil, err
	}

	return res, nil
}

func pareRequestBody(request events.LambdaFunctionURLRequest) (*types.Request, error) {
	requestBody := types.Request{}
	err := json.Unmarshal([]byte(request.Body), &requestBody)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling request body: %v", err)
	}
	return &requestBody, nil
}

func main() {
	lambda.Start(HandleRequest)
}
