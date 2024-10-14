package logger_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/logger"
)

var (
	xmlInput string = `<?xml version="1.0" encoding="utf-8"?>
	<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
		xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
		xmlns:xsd="http://www.w3.org/2001/XMLSchema"
		xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing"
		xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd"
		xmlns:wsu="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd">
		<soap:Header>
			<ConfermaUserStateHeader xmlns="http://cpapi.conferma.com/">
				<AgentID>1234</AgentID>
				<ClientID>4567</ClientID>
				<BookerID>7890</BookerID>
			</ConfermaUserStateHeader>
			<wsa:Action>http://cpapi.conferma.com/GetCardResponse</wsa:Action>
			<wsa:MessageID>urn:uuid:c26698d4-c4b7-4c25-a5a4-38c7346a10ae</wsa:MessageID>
			<wsa:RelatesTo>urn:uuid:cf80673f-a253-4c33-abf4-7ff84473b4fe</wsa:RelatesTo>
			<wsa:To>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</wsa:To>
			<wsse:Security>
				<wsu:Timestamp wsu:Id="Timestamp-8979a216-b4be-4eeb-a423-db4569d492bc">
					<wsu:Created>2021-06-24T12:17:17Z</wsu:Created>
					<wsu:Expires>2021-06-24T12:22:17Z</wsu:Expires>
				</wsu:Timestamp>
			</wsse:Security>
		</soap:Header>
		<soap:Body>
			<GetCardResponse xmlns="http://cpapi.conferma.com/">
				<GetCardResult Type="General" DeploymentID="11181707" CardPoolName="Wego">
					<General>
						<Name>WGML-0019</Name>
						<ConsumerReference>d9x0xzzyzg</ConsumerReference>
						<Amount Value="625.79" Currency="USD" />
						<PaymentRange StartDate="2021-12-02T00:00:00" EndDate="2021-12-05T00:00:00" />
					</General>
					<Card>
						<Name>Wego</Name>
						<Number>4111156600005845</Number>
						<Type>VI</Type>
						<ExpiryDate Month="5" Year="2023" />
						<CVV />
						<Provider ID="52" Name="Ixaris" />
					</Card>
					<Identifiers />
				</GetCardResult>
			</GetCardResponse>
		</soap:Body>
	</soap:Envelope>`
	expectedXMLOutput = `<?xml version="1.0" encoding="utf-8"?>
	<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
		xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
		xmlns:xsd="http://www.w3.org/2001/XMLSchema"
		xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing"
		xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd"
		xmlns:wsu="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd">
		<soap:Header>
			<ConfermaUserStateHeader xmlns="http://cpapi.conferma.com/">
				<AgentID>Wego</AgentID>
				<ClientID>Wego</ClientID>
				<BookerID>Wego</BookerID>
			</ConfermaUserStateHeader>
			<wsa:Action>http://cpapi.conferma.com/GetCardResponse</wsa:Action>
			<wsa:MessageID>urn:uuid:c26698d4-c4b7-4c25-a5a4-38c7346a10ae</wsa:MessageID>
			<wsa:RelatesTo>urn:uuid:cf80673f-a253-4c33-abf4-7ff84473b4fe</wsa:RelatesTo>
			<wsa:To>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</wsa:To>
			<wsse:Security>
				<wsu:Timestamp wsu:Id="Timestamp-8979a216-b4be-4eeb-a423-db4569d492bc">
					<wsu:Created>2021-06-24T12:17:17Z</wsu:Created>
					<wsu:Expires>2021-06-24T12:22:17Z</wsu:Expires>
				</wsu:Timestamp>
			</wsse:Security>
		</soap:Header>
		<soap:Body>
			<GetCardResponse xmlns="http://cpapi.conferma.com/">
				<GetCardResult Type="General" DeploymentID="11181707" CardPoolName="Wego">
					<General>
						<Name>WGML-0019</Name>
						<ConsumerReference>d9x0xzzyzg</ConsumerReference>
						<Amount Value="625.79" Currency="USD" />
						<PaymentRange StartDate="2021-12-02T00:00:00" EndDate="2021-12-05T00:00:00" />
					</General>
					<Card>
						<Name>Wego</Name>
						<Number>Wego</Number>
						<Type>VI</Type>
						<ExpiryDate Month="5" Year="2023" />
						<CVV />
						<Provider ID="52" Name="Ixaris" />
					</Card>
					<Identifiers />
				</GetCardResult>
			</GetCardResponse>
		</soap:Body>
	</soap:Envelope>`
	defaultExpectedXMLOutput = `<?xml version="1.0" encoding="utf-8"?>
	<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
		xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
		xmlns:xsd="http://www.w3.org/2001/XMLSchema"
		xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing"
		xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd"
		xmlns:wsu="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd">
		<soap:Header>
			<ConfermaUserStateHeader xmlns="http://cpapi.conferma.com/">
				<AgentID>[Filtered by Wego]</AgentID>
				<ClientID>[Filtered by Wego]</ClientID>
				<BookerID>[Filtered by Wego]</BookerID>
			</ConfermaUserStateHeader>
			<wsa:Action>http://cpapi.conferma.com/GetCardResponse</wsa:Action>
			<wsa:MessageID>urn:uuid:c26698d4-c4b7-4c25-a5a4-38c7346a10ae</wsa:MessageID>
			<wsa:RelatesTo>urn:uuid:cf80673f-a253-4c33-abf4-7ff84473b4fe</wsa:RelatesTo>
			<wsa:To>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</wsa:To>
			<wsse:Security>
				<wsu:Timestamp wsu:Id="Timestamp-8979a216-b4be-4eeb-a423-db4569d492bc">
					<wsu:Created>2021-06-24T12:17:17Z</wsu:Created>
					<wsu:Expires>2021-06-24T12:22:17Z</wsu:Expires>
				</wsu:Timestamp>
			</wsse:Security>
		</soap:Header>
		<soap:Body>
			<GetCardResponse xmlns="http://cpapi.conferma.com/">
				<GetCardResult Type="General" DeploymentID="11181707" CardPoolName="Wego">
					<General>
						<Name>WGML-0019</Name>
						<ConsumerReference>d9x0xzzyzg</ConsumerReference>
						<Amount Value="625.79" Currency="USD" />
						<PaymentRange StartDate="2021-12-02T00:00:00" EndDate="2021-12-05T00:00:00" />
					</General>
					<Card>
						<Name>Wego</Name>
						<Number>[Filtered by Wego]</Number>
						<Type>VI</Type>
						<ExpiryDate Month="5" Year="2023" />
						<CVV />
						<Provider ID="52" Name="Ixaris" />
					</Card>
					<Identifiers />
				</GetCardResult>
			</GetCardResponse>
		</soap:Body>
	</soap:Envelope>`
	jsonInput = `{
		"test1": [
			{
				"nested": [
					{ "value": "12345" },
					{ "value": "abcde" }
				]
			},
			{
				"nested": [
					{ "value": "54321" },
					{ "value": "edcba" }
				]
			}
		],
		"id": "pay_mbabizu24mvu3mela5njyhpit4",
		"requested_on": "2019-08-24T14:15:22Z",
		"source": {
		  "type": "card",
		  "id": "src_nwd3m4in3hkuddfpjsaevunhdy",
		  "billing_address": {
			"address_line1": "Checkout.com",
			"address_line2": "90 Tottenham Court Road",
			"city": "London",
			"state": "London",
			"zip": "W1T 4TJ",
			"country": "GB"
		  },
		  "phone": {
			"country_code": "+1",
			"number": "4155552671"
		  }
		},
		"destination": {
		  "type": "card",
		  "id": "src_wmlfc3zyhqzehihu7giusaaawu",
		  "billing_address": {
			"address_line1": "Checkout.com",
			"address_line2": "90 Tottenham Court Road",
			"city": "London",
			"state": "London",
			"zip": "W1T 4TJ",
			"country": "GB"
		  },
		  "phone": {
			"country_code": "+1",
			"number": "4155552671"
		  }
		},
		"amount": 6540,
		"currency": "USD",
		"payment_type": "Recurring",
		"reference": "ORD-5023-4E89",
		"description": "Set of 3 masks",
		"approved": true,
		"status": "Authorized",
		"3ds": {
		  "downgraded": false,
		  "enrolled": "Y",
		  "signature_valid": "Y",
		  "authentication_response": "Y",
		  "cryptogram": "hv8mUFzPzRZoCAAAAAEQBDMAAAA=",
		  "xid": "MDAwMDAwMDAwMDAwMDAwMzIyNzY=",
		  "version": "2.1.0",
		  "exemption": "low_value"
		},
		"risk": {
		  "flagged": true
		},
		"customer": {
		  "id": "cus_y3oqhf46pyzuxjbcn2giaqnb44",
		  "email": "jokershere@gmail.com",
		  "name": "Jack Napier"
		},
		"billing_descriptor": {
		  "name": "SUPERHEROES.COM",
		  "city": "GOTHAM"
		},
		"shipping": {
		  "address": {
			"address_line1": "Checkout.com",
			"address_line2": "90 Tottenham Court Road",
			"city": "London",
			"state": "London",
			"zip": "W1T 4TJ",
			"country": "GB"
		  },
		  "phone": {
			"country_code": "+1",
			"number": "4155552671"
		  }
		},
		"payment_ip": "90.197.169.245",
		"recipient": {
		  "dob": "1985-05-15",
		  "account_number": "5555554444",
		  "zip": "W1T",
		  "first_name": "John",
		  "last_name": "Jones",
		  "country": "GB"
		},
		"metadata": {
		  "coupon_code": "NY2018",
		  "partner_id": 123989
		},
		"eci": "06",
		"scheme_id": "488341541494658",
		"actions": [
		  {
			"id": "act_y3oqhf46pyzuxjbcn2giaqnb44",
			"type": "Authorization",
			"response_code": "10000",
			"response_summary": "Approved"
		  }
		],
		"_links": {
		  "self": {
			"href": "https://api.checkout.com/payments/pay_y3oqhf46pyzuxjbcn2giaqnb44"
		  },
		  "actions": {
			"href": "https://api.checkout.com/payments/pay_y3oqhf46pyzuxjbcn2giaqnb44/actions"
		  },
		  "refund": {
			"href": "https://api.checkout.com/payments/pay_y3oqhf46pyzuxjbcn2giaqnb44/refund"
		  }
		}
	  }`
	expectedJSONOutput = `{
		"test1": [
			{
				"nested": [
					{ "value": "Wego" },
					{ "value": "Wego" }
				]
			},
			{
				"nested": [
					{ "value": "Wego" },
					{ "value": "Wego" }
				]
			}
		],
		"id": "Wego",
		"requested_on": "2019-08-24T14:15:22Z",
		"source": {
		  "type": "card",
		  "id": "src_nwd3m4in3hkuddfpjsaevunhdy",
		  "billing_address": {
			"address_line1": "Checkout.com",
			"address_line2": "90 Tottenham Court Road",
			"city": "London",
			"state": "London",
			"zip": "Wego",
			"country": "GB"
		  },
		  "phone": {
			"country_code": "+1",
			"number": "4155552671"
		  }
		},
		"destination": {
		  "type": "card",
		  "id": "src_wmlfc3zyhqzehihu7giusaaawu",
		  "billing_address": {
			"address_line1": "Checkout.com",
			"address_line2": "90 Tottenham Court Road",
			"city": "London",
			"state": "London",
			"zip": "W1T 4TJ",
			"country": "GB"
		  },
		  "phone": {
			"country_code": "+1",
			"number": "4155552671"
		  }
		},
		"amount": 6540,
		"currency": "USD",
		"payment_type": "Recurring",
		"reference": "ORD-5023-4E89",
		"description": "Set of 3 masks",
		"approved": true,
		"status": "Authorized",
		"3ds": {
		  "downgraded": false,
		  "enrolled": "Y",
		  "signature_valid": "Y",
		  "authentication_response": "Y",
		  "cryptogram": "hv8mUFzPzRZoCAAAAAEQBDMAAAA=",
		  "xid": "Wego",
		  "version": "2.1.0",
		  "exemption": "low_value"
		},
		"risk": {
		  "flagged": true
		},
		"customer": {
		  "id": "cus_y3oqhf46pyzuxjbcn2giaqnb44",
		  "email": "jokershere@gmail.com",
		  "name": "Jack Napier"
		},
		"billing_descriptor": {
		  "name": "SUPERHEROES.COM",
		  "city": "GOTHAM"
		},
		"shipping": {
		  "address": {
			"address_line1": "Checkout.com",
			"address_line2": "90 Tottenham Court Road",
			"city": "London",
			"state": "London",
			"zip": "W1T 4TJ",
			"country": "GB"
		  },
		  "phone": {
			"country_code": "+1",
			"number": "4155552671"
		  }
		},
		"payment_ip": "90.197.169.245",
		"recipient": {
		  "dob": "1985-05-15",
		  "account_number": "5555554444",
		  "zip": "W1T",
		  "first_name": "John",
		  "last_name": "Jones",
		  "country": "GB"
		},
		"metadata": {
		  "coupon_code": "NY2018",
		  "partner_id": 123989
		},
		"eci": "06",
		"scheme_id": "488341541494658",
		"actions": [
		  {
			"id": "act_y3oqhf46pyzuxjbcn2giaqnb44",
			"type": "Authorization",
			"response_code": "10000",
			"response_summary": "Approved"
		  }
		],
		"_links": {
		  "self": {
			"href": "https://api.checkout.com/payments/pay_y3oqhf46pyzuxjbcn2giaqnb44"
		  },
		  "actions": {
			"href": "https://api.checkout.com/payments/pay_y3oqhf46pyzuxjbcn2giaqnb44/actions"
		  },
		  "refund": {
			"href": "https://api.checkout.com/payments/pay_y3oqhf46pyzuxjbcn2giaqnb44/refund"
		  }
		}
	  }`
	defaultExpectedJSONOutput = `{
		"test1": [
			{
				"nested": [
					{ "value": "[Filtered by Wego]" },
					{ "value": "[Filtered by Wego]" }
				]
			},
			{
				"nested": [
					{ "value": "[Filtered by Wego]" },
					{ "value": "[Filtered by Wego]" }
				]
			}
		],
		"id": "[Filtered by Wego]",
		"requested_on": "2019-08-24T14:15:22Z",
		"source": {
		  "type": "card",
		  "id": "src_nwd3m4in3hkuddfpjsaevunhdy",
		  "billing_address": {
			"address_line1": "Checkout.com",
			"address_line2": "90 Tottenham Court Road",
			"city": "London",
			"state": "London",
			"zip": "[Filtered by Wego]",
			"country": "GB"
		  },
		  "phone": {
			"country_code": "+1",
			"number": "4155552671"
		  }
		},
		"destination": {
		  "type": "card",
		  "id": "src_wmlfc3zyhqzehihu7giusaaawu",
		  "billing_address": {
			"address_line1": "Checkout.com",
			"address_line2": "90 Tottenham Court Road",
			"city": "London",
			"state": "London",
			"zip": "W1T 4TJ",
			"country": "GB"
		  },
		  "phone": {
			"country_code": "+1",
			"number": "4155552671"
		  }
		},
		"amount": 6540,
		"currency": "USD",
		"payment_type": "Recurring",
		"reference": "ORD-5023-4E89",
		"description": "Set of 3 masks",
		"approved": true,
		"status": "Authorized",
		"3ds": {
		  "downgraded": false,
		  "enrolled": "Y",
		  "signature_valid": "Y",
		  "authentication_response": "Y",
		  "cryptogram": "hv8mUFzPzRZoCAAAAAEQBDMAAAA=",
		  "xid": "[Filtered by Wego]",
		  "version": "2.1.0",
		  "exemption": "low_value"
		},
		"risk": {
		  "flagged": true
		},
		"customer": {
		  "id": "cus_y3oqhf46pyzuxjbcn2giaqnb44",
		  "email": "jokershere@gmail.com",
		  "name": "Jack Napier"
		},
		"billing_descriptor": {
		  "name": "SUPERHEROES.COM",
		  "city": "GOTHAM"
		},
		"shipping": {
		  "address": {
			"address_line1": "Checkout.com",
			"address_line2": "90 Tottenham Court Road",
			"city": "London",
			"state": "London",
			"zip": "W1T 4TJ",
			"country": "GB"
		  },
		  "phone": {
			"country_code": "+1",
			"number": "4155552671"
		  }
		},
		"payment_ip": "90.197.169.245",
		"recipient": {
		  "dob": "1985-05-15",
		  "account_number": "5555554444",
		  "zip": "W1T",
		  "first_name": "John",
		  "last_name": "Jones",
		  "country": "GB"
		},
		"metadata": {
		  "coupon_code": "NY2018",
		  "partner_id": 123989
		},
		"eci": "06",
		"scheme_id": "488341541494658",
		"actions": [
		  {
			"id": "act_y3oqhf46pyzuxjbcn2giaqnb44",
			"type": "Authorization",
			"response_code": "10000",
			"response_summary": "Approved"
		  }
		],
		"_links": {
		  "self": {
			"href": "https://api.checkout.com/payments/pay_y3oqhf46pyzuxjbcn2giaqnb44"
		  },
		  "actions": {
			"href": "https://api.checkout.com/payments/pay_y3oqhf46pyzuxjbcn2giaqnb44/actions"
		  },
		  "refund": {
			"href": "https://api.checkout.com/payments/pay_y3oqhf46pyzuxjbcn2giaqnb44/refund"
		  }
		}
	  }`
)

func Test_RedactXML_ReturnErrorText_WithInvalidInput(t *testing.T) {
	assert := assert.New(t)

	output := logger.RedactXML("<wego", "", []string{})
	assert.Contains(output, "invalid XML input")
}

func Test_RedactXML_DoNothing_WhenInputEmptyTags(t *testing.T) {
	assert := assert.New(t)

	output := logger.RedactXML(xmlInput, "", []string{})
	assert.Equal(xmlInput, output)
}

func Test_RedactXML_DoNothing_WhenInputDoesNotContainTags(t *testing.T) {
	assert := assert.New(t)

	output := logger.RedactXML(xmlInput, "", []string{"wego"})
	assert.Equal(xmlInput, output)
}

func Test_RedactXML_Ok(t *testing.T) {
	assert := assert.New(t)

	output := logger.RedactXML(xmlInput, "Wego", []string{"AgentID", "ClientID", "BookerID", "Number"})
	assert.Equal(expectedXMLOutput, output)

	output = logger.RedactXML(xmlInput, "", []string{"AgentID", "ClientID", "BookerID", "Number"})
	assert.Equal(defaultExpectedXMLOutput, output)
}

func Test_RedactJSON_InvalidInput(t *testing.T) {
	assert := assert.New(t)

	output := logger.RedactJSON(xmlInput, "Wego", [][]string{{"id"}, {"source", "billing_address", "zip"}, {"3ds", "xid"}})
	assert.Contains(output, "cannot parse JSON")
}

func Test_RedactJSON_DoNothing_WhenKeysNotFound(t *testing.T) {
	assert := assert.New(t)
	var compactOutput, compactExpectedOutput bytes.Buffer

	// not provide keys
	output := logger.RedactJSON(jsonInput, "", nil)
	err := json.Compact(&compactOutput, []byte(output))
	assert.NoError(err)
	err = json.Compact(&compactExpectedOutput, []byte(jsonInput))
	assert.NoError(err)
	assert.Equal(compactExpectedOutput, compactOutput)

	// empty keys
	output = logger.RedactJSON(jsonInput, "", [][]string{})
	err = json.Compact(&compactOutput, []byte(output))
	assert.NoError(err)
	err = json.Compact(&compactExpectedOutput, []byte(jsonInput))
	assert.NoError(err)
	assert.Equal(compactExpectedOutput, compactOutput)

	// keys not exist
	output = logger.RedactJSON(jsonInput, "Wego", [][]string{{"yo"}, {"source", "billing_address", "whatsup"}, {"3ds", "hi"}})
	err = json.Compact(&compactOutput, []byte(output))
	assert.NoError(err)
	err = json.Compact(&compactExpectedOutput, []byte(jsonInput))
	assert.NoError(err)
	assert.Equal(compactExpectedOutput, compactOutput)
}

func Test_RedactJSON_Ok(t *testing.T) {
	assert := assert.New(t)
	var compactOutput, compactExpectedOutput bytes.Buffer

	output := logger.RedactJSON(jsonInput, "Wego", [][]string{
		{"id"},
		{"source", "billing_address", "zip"},
		{"3ds", "xid"},
		{"test1", "[]", "nested", "[]", "value"},
	})

	err := json.Compact(&compactOutput, []byte(output))
	assert.NoError(err)
	err = json.Compact(&compactExpectedOutput, []byte(expectedJSONOutput))
	assert.NoError(err)
	assert.Equal(compactExpectedOutput, compactOutput)

	output = logger.RedactJSON(jsonInput, "", [][]string{
		{"id"},
		{"source", "billing_address", "zip"},
		{"3ds", "xid"},
		{"test1", "[]", "nested", "[]", "value"},
	})
	err = json.Compact(&compactOutput, []byte(output))
	assert.NoError(err)
	err = json.Compact(&compactExpectedOutput, []byte(defaultExpectedJSONOutput))
	assert.NoError(err)
	assert.Equal(compactExpectedOutput, compactOutput)
}

func BenchmarkRedactJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = logger.RedactJSON(jsonInput, "[Filtered by Wego]", [][]string{
			{"id"},
			{"source", "billing_address", "zip"},
			{"3ds", "xid"},
			{"test1", "[]", "nested", "[]", "value"},
		})
	}
}

func BenchmarkRedactJSONParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = logger.RedactJSON(jsonInput, "[Filtered by Wego]", [][]string{
				{"id"},
				{"source", "billing_address", "zip"},
				{"3ds", "xid"},
				{"test1", "[]", "nested", "[]", "value"},
			})
		}
	})
}
