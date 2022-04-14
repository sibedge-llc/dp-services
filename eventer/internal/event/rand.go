package event

import (
	"encoding/json"

	"github.com/bxcodec/faker/v3"

	browser "github.com/EDDYCJY/fake-useragent"
)

type SomeStructWithTags struct {
	Latitude           float32 `faker:"lat" json:"lat"`
	Longitude          float32 `faker:"long" json:"long"`
	CreditCardNumber   string  `faker:"cc_number" json:"cc_number"`
	CreditCardType     string  `faker:"cc_type" json:"cc_type"`
	Email              string  `faker:"email" json:"email"`
	DomainName         string  `faker:"domain_name" json:"domain_name"`
	IPV4               string  `faker:"ipv4" json:"ipv4"`
	IPV6               string  `faker:"ipv6" json:"ipv6"`
	Password           string  `faker:"password" json:"password"`
	Jwt                string  `faker:"jwt" json:"jwt"`
	PhoneNumber        string  `faker:"phone_number" json:"phone_number"`
	MacAddress         string  `faker:"mac_address" json:"mac_address"`
	URL                string  `faker:"url" json:"url"`
	UserName           string  `faker:"username" json:"username"`
	TollFreeNumber     string  `faker:"toll_free_number" json:"toll_free_number"`
	E164PhoneNumber    string  `faker:"e_164_phone_number" json:"e_164_phone_number"`
	FirstName          string  `faker:"first_name" json:"first_name"`
	LastName           string  `faker:"last_name" json:"last_name"`
	Name               string  `faker:"name" json:"name"`
	UnixTime           int64   `faker:"unix_time" json:"unix_time"`
	Date               string  `faker:"date" json:"date"`
	Time               string  `faker:"time" json:"time"`
	MonthName          string  `faker:"month_name" json:"month_name"`
	Year               string  `faker:"year" json:"year"`
	DayOfWeek          string  `faker:"day_of_week" json:"day_of_week"`
	DayOfMonth         string  `faker:"day_of_month" json:""`
	Timestamp          string  `faker:"timestamp" json:"timestamp"`
	Century            string  `faker:"century" json:"century"`
	TimeZone           string  `faker:"timezone" json:"timezone"`
	TimePeriod         string  `faker:"time_period" json:"time_period"`
	Word               string  `faker:"word" json:"word"`
	Sentence           string  `faker:"sentence" json:"sentence"`
	Paragraph          string  `faker:"paragraph" json:"paragraph"`
	Currency           string  `faker:"currency" json:"currency"`
	Amount             float64 `faker:"amount" json:"amount"`
	AmountWithCurrency string  `faker:"amount_with_currency" json:"amount_with_currency"`
	UUIDHypenated      string  `faker:"uuid_hyphenated" json:"uuid_hyphenated"`
	UUID               string  `faker:"uuid_digit" json:"uuid_digit"`
	PaymentMethod      string  `faker:"oneof: cc, paypal, check, money order" json:"payment_method"`
	ID                 int64   `faker:"oneof: 1, 10000" json:"id"`
	Price              float64 `faker:"oneof: 1.5, 100.99" json:"price"`
	Number             int64   `faker:"oneof: 1, 10000" json:"number"`
}

func getRandData() (interface{}, error) {
	var s SomeStructWithTags
	err := faker.FakeData(&s)
	if err != nil {
		return map[string]interface{}{}, err
	}
	b, err := json.Marshal(s)
	if err != nil {
		return map[string]interface{}{}, err
	}
	var r map[string]interface{}
	err = json.Unmarshal(b, &r)
	if err != nil {
		return map[string]interface{}{}, err
	}
	return r, nil
}

func getRandUserAgent() string {
	return browser.Random()
}
