package hydra

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type HydraRepository struct {
	store *Store
}

func InitRepository(store *Store) *HydraRepository {
	return &HydraRepository{store: store}
}

// Получение пользователя гидры по customer code
func (r *HydraRepository) GetCustomerByCode(code string) (*HydraCustomer, error) {
	var user HydraCustomer

	err := r.store.db.QueryRow(
		`SELECT si_v_user_contracts.n_subject_id, si_v_subj_accounts.vc_subj_code, si_v_subj_accounts.n_account_id, si_v_user_contracts.n_doc_id, si_v_user_devices_simple.n_device_id, si_v_subj_accounts.VC_ACCOUNT
		   FROM SI_V_SUBJ_ACCOUNTS 
		   LEFT JOIN SI_V_USER_CONTRACTS ON si_v_subj_accounts.n_subject_id = SI_V_USER_CONTRACTS.n_subject_id
		   LEFT JOIN SI_V_USER_DEVICES_SIMPLE on si_v_subj_accounts.n_subject_id = SI_V_USER_DEVICES_SIMPLE.N_SUBJECT_ID
		   WHERE si_v_subj_accounts.VC_ACCOUNT = :1 AND SI_V_USER_DEVICES_SIMPLE.N_DEVICE_GOOD_ID = 21501
`, code).Scan(&user.CustomerId, &user.CustomerLogin, &user.AccountId, &user.DocId, &user.DeviceId, &user.CustomerCode)

	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("Абонент не найден")
		}
		return nil, err
	}

	if user.DeviceId == "" {
		err = fmt.Errorf("Не найдено оконечное оборудование")
		return nil, err
	}

	return &user, nil
}

// Получение пользователя гидры по customer id
func (r *HydraRepository) GetCustomerById(id string) (*HydraCustomer, error) {
	var user HydraCustomer

	err := r.store.db.QueryRow(
		`SELECT si_v_user_contracts.n_subject_id, si_v_subj_accounts.vc_subj_code, si_v_subj_accounts.n_account_id, si_v_user_contracts.n_doc_id, si_v_user_devices_simple.n_device_id
		   FROM SI_V_SUBJ_ACCOUNTS 
		   LEFT JOIN SI_V_USER_CONTRACTS ON si_v_subj_accounts.n_subject_id = SI_V_USER_CONTRACTS.n_subject_id
		   LEFT JOIN SI_V_USER_DEVICES_SIMPLE on si_v_subj_accounts.n_subject_id = SI_V_USER_DEVICES_SIMPLE.N_SUBJECT_ID
		   WHERE SI_V_SUBJ_ACCOUNTS.n_subject_id = :1 AND SI_V_USER_DEVICES_SIMPLE.N_DEVICE_GOOD_ID = 21501`, id).Scan(&user.CustomerId, &user.CustomerCode, &user.AccountId, &user.DocId, &user.DeviceId)

	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("Абонент не найден")
		}
		return nil, err
	}

	if user.DeviceId == "" {
		err = fmt.Errorf("Не найдено оконечное оборудование")
		return nil, err
	}

	return &user, nil
}

// Получение активного акта начисления по account id
func (r *HydraRepository) GetChargeLogByAccountId(id string) (*HydraChargeLog, error) {
	var chargelog HydraChargeLog
	query := "SELECT N_DOC_ID, VC_NAME FROM SD_V_CHARGE_LOGS where n_account_id = :1 AND (N_DOC_STATE_ID = 4003 OR N_DOC_STATE_ID=10003)"

	err := r.store.db.QueryRow(query, id).Scan(&chargelog.DocId, &chargelog.Name)

	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("У абонента отсутствуют активные подписки")
		}
		return nil, err
	}
	return &chargelog, nil
}

// Закрытие акта начислений
func (r *HydraRepository) CloseChargeLog(docId string) error {

	currentTime := time.Now().Format("02.01.2006 15:04:05")
	query :=
		`
			BEGIN
			  SI_USERS_PKG.CHANGE_CHARGE_LOG_PERIOD(
				num_N_DOC_ID => :1,
				dt_D_OPER    => TO_DATE(:2, 'DD.MM.YYYY HH24:MI:SS'));
			COMMIT;
			END;
	`
	_, err := r.store.db.Exec(query, docId, currentTime)
	if err != nil {
		return err
	}

	return nil
}

// Получить активную блокировку пользователя
func (r *HydraRepository) GetCurrentBlockSubscription(accountId string) (*HydraSubscrition, error) {
	var subs HydraSubscrition
	query := "SELECT N_SUBJ_GOOD_ID FROM SI_V_USER_GOODS WHERE N_ACCOUNT_ID = :1 AND N_DOC_STATE_ID = 4003 AND C_FL_CLOSED = 'N' AND N_GOOD_ID = :2"

	err := r.store.db.QueryRow(query, accountId, r.store.config.HydraBlockSubscritionId).Scan(&subs.SubjGoodId)

	if err != nil {
		return nil, err
	}

	return &subs, nil

}

func (r *HydraRepository) SetBlock(customer *HydraCustomer) error {
	query := `
		DECLARE
		num_N_SUBJ_GOOD_ID                  SI_V_USER_GOODS.N_SUBJ_GOOD_ID%TYPE;
		BEGIN
		SI_USERS_PKG.SI_USER_GOODS_PUT(
			num_N_SUBJ_GOOD_ID      => num_N_SUBJ_GOOD_ID,
			num_N_GOOD_ID           => :1,
			num_N_SUBJECT_ID        => :2,
			num_N_ACCOUNT_ID        => :3,
			num_N_OBJECT_ID         => :4,
			num_N_PAY_DAY           => 1,
			num_N_LINE_NO           => 1000000,
			num_N_DOC_ID            => :5,
			num_N_UNIT_ID           => SYS_CONTEXT('CONST', 'UNIT_Unknown'));
	
		COMMIT;
		END;
	`

	_, err := r.store.db.Exec(query, r.store.config.HydraBlockSubscritionId, customer.CustomerId, customer.AccountId, customer.DeviceId, customer.DocId)

	if err != nil {
		return err
	}

	return nil
}

func (r *HydraRepository) CloseBlock(sub *HydraSubscrition) error {
	currentTime := time.Now().Format("02.01.2006 15:04:05")
	fmt.Println(currentTime)
	query := `
	BEGIN
	  SI_USERS_PKG.SI_USER_GOODS_CLOSE(
		num_N_SUBJ_GOOD_ID => :1,
		dt_D_END           => TO_DATE(:2, 'DD.MM.YYYY HH24:MI:SS'));
	COMMIT;
	END;
`
	var err error
	if sub.SubjGoodId != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, err = r.store.db.QueryContext(ctx, query, sub.SubjGoodId, currentTime)
	}

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (r *HydraRepository) GetChargeLogServices(chargeLogId string) (*[]HydraChargeLog, error) {
	var chargeLogs []HydraChargeLog
	query := `
      SELECT N_DOC_ID, VC_GOOD_NAME, N_GOOD_ID
	  FROM   SD_V_CHARGE_LOGS_C
	  WHERE  N_DOC_ID = :1`

	rows, err := r.store.db.Query(query, chargeLogId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var chargeLog HydraChargeLog
		err := rows.Scan(&chargeLog.DocId, &chargeLog.Name, &chargeLog.GoodId)
		if err != nil {
			return nil, err
		}
		chargeLogs = append(chargeLogs, chargeLog)
	}

	return &chargeLogs, nil
}

// Выставление акта начислений
func (r *HydraRepository) ChargingChargeLog(customerId string) error {
	fmt.Println("SETTING CHARGE LOG")
	fmt.Println(customerId)
	query := `BEGIN
		  SD_CHARGE_LOGS_CHARGING_PKG.PROCESS_SUBJECT(
		  num_N_SUBJECT_ID     => :1);
		  COMMIT;
		END;`
	_, err := r.store.db.Exec(query, customerId)
	if err != nil {
		return err
	}

	return nil
}

func (r *HydraRepository) GetPreparedChargeLog(customerId string) (*HydraChargeLog, error) {
	return nil, nil
}
