package params

import "testing"

func TestValidateJSONAcceptsQDM12CompatibleConfigs(t *testing.T) {
	t.Parallel()

	testCases := map[string]string{
		"netcup": `{
			"settings": [{
				"provider": "netcup",
				"domain": "sub.example.com",
				"api_key": "api-key",
				"password": "api-password",
				"customer_number": "123456",
				"ip_version": "ipv4"
			}]
		}`,
		"multiple providers": `{
			"settings": [
				{
					"provider": "namecheap",
					"domain": "example.com",
					"password": "e5322165c1d74692bfa6d807100c0310"
				},
				{
					"provider": "duckdns",
					"domain": "example.duckdns.org",
					"token": "00000000-0000-0000-0000-000000000000"
				},
				{
					"provider": "godaddy",
					"domain": "subdomain.example.org",
					"key": "dLP4WKz5PdkS_GuUDNigHcLQFpw4CWNwAQ5",
					"secret": "GuUFdVFj8nJ1M79RtdwmkZ"
				}
			]
		}`,
		"legacy host owner": `{
			"settings": [{
				"provider": "duckdns",
				"domain": "example.duckdns.org",
				"host": "@",
				"token": "00000000-0000-0000-0000-000000000000"
			}]
		}`,
		"dyn password retrocompatibility": `{
			"settings": [{
				"provider": "dyn",
				"domain": "example.com",
				"username": "username",
				"password": "legacy-client-key",
				"ip_version": "ipv4"
			}]
		}`,
	}

	for name, configJSON := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := ValidateJSON([]byte(configJSON))
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestValidateJSONRejectsBrokenConfig(t *testing.T) {
	t.Parallel()

	const configJSON = `{
		"settings": [{
			"provider": "netcup",
			"domain": "example.com",
			"api_key": "api-key",
			"password": "api-password"
		}]
	}`

	_, err := ValidateJSON([]byte(configJSON))
	if err == nil {
		t.Fatal("expected missing customer_number to be rejected")
	}
}
