package validators

import (
	"cf-oidc-proxy/config"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"strings"
)

func GroupEntityMatcher(cfg config.RoleConfig, groupName string, idToken *oidc.IDToken) (match bool, err error) {
	group, err := cfg.GetRoleByName(groupName)
	if err != nil {
		return false, err
	}
	matchers := group.Entities.Matchers
	claims := map[string]string{}
	idToken.Claims(&claims)

	for _, matcher := range matchers {
		expectedClaims := matcher.Claims
		for key, value := range expectedClaims {
			match := ExecuteOperator(matcher.Operator, value, claims[key])
			fmt.Println("key:", key, "expected;", value, "actual:", claims[key], "match:", match)
			if match != true {
				return false, nil
			}
		}
	}
	return true, nil
}

func ExecuteOperator(operator string, expected string, actual string) (match bool) {
	switch operator {
	case "StringEquals":
		return StringEquals(expected, actual)
	case "StringNotEquals":
		return StringNotEquals(expected, actual)
	case "StringEqualsIgnoreCase":
		return StringEqualsIgnoreCase(expected, actual)
	case "StringNotEqualsIgnoreCase":
		return StringNotEqualsIgnoreCase(expected, actual)
	}
	return false
}

func StringEquals(expect string, actual string) bool {
	if strings.Contains(expect, "*") {
		return wildcardMatch(expect, actual)
	} else {
		return expect == actual
	}
}

func StringEqualsIgnoreCase(expect string, actual string) bool {
	return StringEquals(strings.ToLower(expect), strings.ToLower(actual))
}

func StringNotEqualsIgnoreCase(expect string, actual string) bool {
	return !StringEqualsIgnoreCase(expect, actual)
}

func StringNotEquals(expect string, actual string) bool {
	return !StringEquals(expect, actual)
}

func wildcardMatch(pattern string, value string) bool {
	// Split the pattern into segments separated by the wildcard (*)
	segments := strings.Split(pattern, "*")
	if len(segments) == 1 {
		// Pattern does not contain a wildcard, so do a regular string comparison
		return pattern == value
	}

	// Check if the value starts with the first segment of the pattern
	if !strings.HasPrefix(value, segments[0]) {
		return false
	}

	// Check if the value ends with the last segment of the pattern
	if !strings.HasSuffix(value, segments[len(segments)-1]) {
		return false
	}

	// Check if the remaining segments of the pattern occur in the value in order
	start := strings.Index(value, segments[0]) + len(segments[0])
	for _, segment := range segments[1 : len(segments)-1] {
		end := strings.Index(value[start:], segment)
		if end == -1 {
			return false
		}
		start += end + len(segment)
	}
	return true
}
