package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"hydra-blocking/external/utils"
	"log"
	"net/http"
	"time"
)

type Error struct {
	Error     string `json:"error"`
	LastBlock string `json:"lastBlock"`
}

type Status struct {
	Status string `json:"status"`
}

func (s *Server) SetBlockHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rq := struct {
		CustomerCode string `json:"customer_code"`
	}{}

	dec := json.NewDecoder(r.Body)

	err := dec.Decode(&rq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		er := Error{Error: "Ошибка в запросе"}
		resp, _ := json.Marshal(er)
		w.Write(resp)
		return
	}

	customer, err := s.hydraStore.Repository.GetCustomerByCode(rq.CustomerCode)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		er := Error{Error: "Ошибка при получении данных пользователя"}
		resp, _ := json.Marshal(er)
		w.Write(resp)
		return
	}

	// Получаем последнюю блокировку аккаунта
	block, err := s.localStore.Repository.GetLatestBlock(customer)
	if err != nil {
		log.Println(err)
		return
	}

	// Проверям, прошло ли 10 дней
	nowTime := time.Now()
	diff := nowTime.Sub(block.EndDate.Time)
	deltaDays := diff.Hours() / 24
	if deltaDays <= 10 {
		w.WriteHeader(http.StatusOK)
		lastBlockTime := block.EndDate.Time.Format("02.01.2006")
		er := Error{Error: "С последней блокировки не прошло 10 дней", LastBlock: lastBlockTime}
		resp, _ := json.Marshal(er)
		w.Write(resp)
		return
	}

	// Подключаем услугу блокировки
	err = s.hydraStore.Repository.SetBlock(customer)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		er := Error{Error: "Ошибка при подключении услуги"}
		resp, _ := json.Marshal(er)
		w.Write(resp)
		return
	}

	// Получаем текущий акт начисления
	chargeLog, err := s.hydraStore.Repository.GetChargeLogByAccountId(customer)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		er := Error{Error: "Ошибка при получении активного акта начислений"}
		resp, _ := json.Marshal(er)
		w.Write(resp)
		return
	}

	// Закрываем акт начисления
	err = s.hydraStore.Repository.CloseChargeLog(chargeLog.DocId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		er := Error{Error: "Ошибка при закрытии действующего акта начисления"}
		resp, _ := json.Marshal(er)
		w.Write(resp)
		return
	}

	// Добавляем блокировку в локальную БД
	err = s.localStore.Repository.CreateBlock(customer)
	if err != nil {
		log.Println("Ошибка при добавлении в локальную БД", err)
	}

	// Выставляем новый акт начислений
	err = s.hydraStore.Repository.ChargingChargeLog(customer.CustomerId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusOK)
		er := Error{Error: "Ошибка при выставлении акта начислений"}
		resp, _ := json.Marshal(er)
		w.Write(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
	status := Status{Status: "OK"}
	resp, _ := json.Marshal(status)
	w.Write(resp)
}

func (s *Server) RemoveBlockHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rq := struct {
		CustomerCode string `json:"customer_code"`
	}{}

	dec := json.NewDecoder(r.Body)

	err := dec.Decode(&rq)
	if err != nil || rq.CustomerCode == "" {
		w.WriteHeader(http.StatusBadRequest)
		er := Error{Error: "Ошибка в запросе"}
		resp, _ := json.Marshal(er)
		w.Write(resp)
		return
	}

	// Получаем пользователя
	customer, err := s.hydraStore.Repository.GetCustomerByCode(rq.CustomerCode)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		er := Error{Error: "Ошибка при получении абонента"}
		resp, _ := json.Marshal(er)
		w.Write(resp)
		return
	}

	// Получаем текущую добровольную блокировку
	currentBlock, err := s.hydraStore.Repository.GetCurrentBlockSubscription(customer.AccountId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		er := Error{Error: "Ошибка при получении текущей блокировки"}
		resp, _ := json.Marshal(er)
		w.Write(resp)
		return
	}

	// Закрываем услугу блокировки
	err = s.hydraStore.Repository.CloseBlock(currentBlock)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		er := Error{Error: "Ошибка при закрытии услуги блокировки"}
		resp, _ := json.Marshal(er)
		w.Write(resp)
		return
	}

	// Получаем текущий Акт Начислений
	chargeLog, err := s.hydraStore.Repository.GetChargeLogByAccountId(customer)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		er := Error{Error: "Ошибка при получении акта начислений"}
		resp, _ := json.Marshal(er)
		w.Write(resp)
		return
	}

	// Закрываем акт начислений
	err = s.hydraStore.Repository.CloseChargeLog(chargeLog.DocId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		er := Error{Error: "Ошибка при закрытии акта начислений"}
		resp, _ := json.Marshal(er)
		w.Write(resp)
		return
	}

	// Ставим дату окончания в локальную базу
	err = s.localStore.Repository.UpdateLatestBlock(customer)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusOK)
		er := Error{Error: "Ошибка при обновлении в локальной базе данных"}
		resp, _ := json.Marshal(er)
		w.Write(resp)
		return
	}

	// Выставляем новый акт начислений
	err = s.hydraStore.Repository.ChargingChargeLog(customer.CustomerId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusOK)
		er := Error{Error: "Ошибка при выставлении акта начислений"}
		resp, _ := json.Marshal(er)
		w.Write(resp)
		return
	}
	w.WriteHeader(http.StatusOK)
	status := Status{Status: "OK"}
	resp, _ := json.Marshal(status)
	w.Write(resp)
}

func (s *Server) GetStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	rq := struct {
		CustomerCode string `json:"customer_code"`
	}{}

	dec := json.NewDecoder(r.Body)

	err := dec.Decode(&rq)

	if err != nil || rq.CustomerCode == "" {
		w.WriteHeader(http.StatusBadRequest)
		er := Error{Error: "Ошибка в запросе"}
		resp, _ := json.Marshal(er)
		w.Write(resp)
		return
	}

	// Получаем пользователя
	customer, err := s.hydraStore.Repository.GetCustomerByCode(rq.CustomerCode)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		er := Error{Error: "Ошибка при получении абонента"}
		resp, _ := json.Marshal(er)
		w.Write(resp)
		return
	}

	// Получаем текущий акт начисления
	chargeLog, err := s.hydraStore.Repository.GetChargeLogByAccountId(customer)
	fmt.Println(chargeLog)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		er := Error{Error: "Ошибка при получении акта начислений"}
		resp, _ := json.Marshal(er)
		w.Write(resp)
		return
	}

	// Получаем все услуги из текущего акта начислений
	services, err := s.hydraStore.Repository.GetChargeLogServices(chargeLog.DocId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		er := Error{Error: "Ошибка при получении активных услуг"}
		resp, _ := json.Marshal(er)
		w.Write(resp)
		return
	}

	fmt.Println(len(*services), "COUNT OF SERVBICES")
	// Если массив больше чем 0
	if len(*services) > 0 {
		for _, service := range *services {
			if service.GoodId == s.config.HydraBlockSubscritionId {
				w.WriteHeader(http.StatusOK)
				status := Status{Status: "blocked"}
				resp, _ := json.Marshal(status)
				w.Write(resp)
				return
			} else {
				w.WriteHeader(http.StatusOK)
				status := Status{Status: "active"}
				resp, _ := json.Marshal(status)
				w.Write(resp)
				return
			}
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		er := Error{Error: "Нет активных услуг в акте начислений"}
		resp, _ := json.Marshal(er)
		w.Write(resp)
		log.Println("Нет активных услуг в акте начислений")
		return
	}

}

func (s *Server) MainHandler(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		CustomerCode string
		Hash         string
		Error        string
	}

	// Parse GET params
	customerCode := r.URL.Query().Get("customer_code")
	requestHash := r.URL.Query().Get("hash")

	// Forming template data
	resp := Response{
		CustomerCode: customerCode,
		Hash:         requestHash,
	}

	if customerCode == "" {
		resp.Error = "Не получены необходимые аттрибуты"
	}

	validateHash := utils.ValidateHash(customerCode, requestHash, s.config)
	if validateHash == false {
		log.Println("ERROR VALIDATING HASH")
		return
	}

	// Render HTML Template
	t, _ := template.ParseFiles("./templates/index.gohtml", "./templates/base.gohtml")
	err := t.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
}
