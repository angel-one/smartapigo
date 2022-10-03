package smartapigo

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	httpmock "github.com/jarcoal/httpmock"
)

// Test New Smart API Connect instance
func TestNewClient(t *testing.T) {
	t.Parallel()

	clientcode := "test"
	password := "test@444"
	apiKey := "test_key"
	totp := "12345"
	client := New(clientcode, password, apiKey, totp)

	if client.password != password || client.clientCode != clientcode || totp != client.totp {
		t.Errorf("Credentials not assigned properly.")
	}
}

// Test all client setters
func TestClientSetters(t *testing.T) {
	t.Parallel()

	clientcode := "test"
	password := "test@444"
	apiKey := "test_key"
	totp := "123456"
	client := New(clientcode, password, apiKey, totp)

	customDebug := true
	customBaseURI := "test"
	customTimeout := 1000 * time.Millisecond
	customAccessToken := "accesstoken"
	customHTTPClientTimeout := time.Duration(2000)
	customHTTPClient := &http.Client{
		Timeout: customHTTPClientTimeout,
	}

	// Check if default debug is false
	if client.debug != false || client.httpClient.GetClient().debug != false {
		t.Errorf("Default debug is not false.")
	}

	// Set custom debug
	client.SetDebug(customDebug)
	if client.debug != customDebug || client.httpClient.GetClient().debug != customDebug {
		t.Errorf("Debug is not set properly.")
	}

	// Test default base uri
	if client.baseURI != baseURI {
		t.Errorf("Default base URI is not set properly.")
	}

	// Set custom base URI
	client.SetBaseURI(customBaseURI)
	if client.baseURI != customBaseURI {
		t.Errorf("Base URI is not set properly.")
	}

	// Test default timeout
	if client.httpClient.GetClient().client.Timeout != requestTimeout {
		t.Errorf("Default request timeout is not set properly.")
	}

	// Set custom timeout for default http client
	client.SetTimeout(customTimeout)
	if client.httpClient.GetClient().client.Timeout != customTimeout {
		t.Errorf("HTTPClient timeout is not set properly.")
	}

	// Set access token
	client.SetAccessToken(customAccessToken)
	if client.accessToken != customAccessToken {
		t.Errorf("Access token is not set properly.")
	}

	// Set custom HTTP Client
	client.SetHTTPClient(customHTTPClient)
	if client.httpClient.GetClient().client != customHTTPClient {
		t.Errorf("Custom HTTPClient is not set properly.")
	}

	// Set timeout for custom http client
	if client.httpClient.GetClient().client.Timeout != customHTTPClientTimeout {
		t.Errorf("Custom HTTPClient timeout is not set properly.")
	}

	// Set custom timeout for custom http client
	client.SetTimeout(customTimeout)
	if client.httpClient.GetClient().client.Timeout != customTimeout {
		t.Errorf("HTTPClient timeout is not set properly.")
	}
}

// Following boiler plate is used to implement setup/teardown using Go subtests feature
const mockBaseDir = "./mock_responses"

var MockResponders = [][]string{
	// Array of [<httpMethod>, <url>, <file_name>]

	// GET endpoints
	{http.MethodGet, URIUserProfile, "profile.json"},
	{http.MethodGet, URIGetPositions, "positions.json"},
	{http.MethodGet, URIGetHoldings, "holdings.json"},
	{http.MethodGet, URIRMS, "rms.json"},
	{http.MethodGet, URIGetTradeBook, "trades.json"},
	{http.MethodGet, URIGetOrderBook, "orders.json"},

	// POST endpoints
	{http.MethodPost, URIModifyOrder, "order_response.json"},
	{http.MethodPost, URIPlaceOrder, "order_response.json"},
	{http.MethodPost, URICancelOrder, "order_response.json"},
	{http.MethodPost, URILTP, "ltp.json"},
	{http.MethodPost, URILogin, "session.json"},
	{http.MethodPost, URIUserSessionRenew, "session.json"},
	{http.MethodPost, URIUserProfile, "profile.json"},
	{http.MethodPost, URILogout, "logout.json"},
	{http.MethodPost, URIConvertPosition, "position_conversion.json"},
}

// Test only function prefix with this
const suiteTestMethodPrefix = "Test"

// TestSuite is an interface where you define suite and test case preparation and tear down logic.
type TestSuite struct {
	TestConnect *Client
}

// Setup the API suit
func (ts *TestSuite) SetupAPITestSuit() {

	clientcode := "test"
	password := "test@444"
	apiKey := "test_key"
	totp := "123456"
	ts.TestConnect = New(clientcode, password, apiKey, totp)
	httpmock.ActivateNonDefault(ts.TestConnect.httpClient.GetClient().client)

	for _, v := range MockResponders {
		httpMethod := v[0]
		route := v[1]
		filePath := v[2]

		resp, err := ioutil.ReadFile(path.Join(mockBaseDir, filePath))
		if err != nil {
			panic("Error while reading mock response: " + filePath)
		}

		base, err := url.Parse(ts.TestConnect.baseURI)
		if err != nil {
			panic("Something went wrong")
		}
		// Replace all url variables with string "test"
		re := regexp.MustCompile("%s")
		formattedRoute := re.ReplaceAllString(route, "test")
		base.Path = path.Join(base.Path, formattedRoute)
		// fmt.Println(base.String())
		// endpoint := path.Join(ts.KiteConnect.baseURI, route)
		httpmock.RegisterResponder(httpMethod, base.String(), httpmock.NewBytesResponder(200, resp))

	}
}

// TearDown API suit
func (ts *TestSuite) TearDownAPITestSuit() {
	//defer httpmock.DeactivateAndReset()
}

// Individual test setup
func (ts *TestSuite) SetupAPITest() {}

// Individual test teardown
func (ts *TestSuite) TearDownAPITest() {}

/*
Run sets up the suite, runs its test cases and tears it down:
 1. Calls `ts.SetUpSuite`
 2. Seeks for any methods that have `Test` prefix, for each of them it:
    a. Calls `SetUp`
    b. Calls the test method itself
    c. Calls `TearDown`
 3. Calls `ts.TearDownSuite`
*/
func RunAPITests(t *testing.T, ts *TestSuite) {
	ts.SetupAPITestSuit()

	suiteType := reflect.TypeOf(ts)
	for i := 0; i < suiteType.NumMethod(); i++ {
		m := suiteType.Method(i)
		if strings.HasPrefix(m.Name, suiteTestMethodPrefix) {
			t.Run(m.Name, func(t *testing.T) {
				ts.SetupAPITest()
				defer ts.TearDownAPITest()

				in := []reflect.Value{reflect.ValueOf(ts), reflect.ValueOf(t)}
				m.Func.Call(in)
			})
		}
	}
}

func TestAPIMethods(t *testing.T) {
	s := &TestSuite{}
	RunAPITests(t, s)
}
