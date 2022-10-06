package hydra

type HydraCustomer struct {
	CustomerId    string `json:"customer_id"`
	CustomerLogin string `json:"customer_login"`
	CustomerCode  string `json:"customer_code"`
	AccountId     string `json:"account_id"`
	DocId         string `json:"doc_id"`
	DeviceId      string `json:"device_id"`
}

type HydraChargeLog struct {
	DocId   string `json:"doc_id"`
	Name    string `json:"name"`
	GoodId  string `json:"good_id"`
	StateId string `json:"state_id"`
}

type HydraSubscrition struct {
	SubjGoodId string `json:"subj_good_id"`
}
