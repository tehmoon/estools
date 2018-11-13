package alert

//import (
//	"time"
//	"../util"
//	"github.com/tehmoon/errors"
//)
//
//type AlertTrigger struct {
//	TriggeredAt time.Time
//	config *TriggerConfig
//}
//
//func (at *AlertTrigger) Increment(config *TriggerConfig) {
////	at.RemainingAlerts++
//
//	if at.config == nil {
////		at.RemainingAlerts = 0
//		at.config = config
//	}
//}
//
//func (at *AlertTrigger) Flush(alert bool) {
//	if at.config == nil {
//		return
//	}
//
//	err := at.config.Alert.Trigger(alert, at.config.PublicURL)
//	if err != nil {
//		err = errors.Wrapf(err, "Error triggering alert")
//		util.Println(err.Error())
//	}
//
//	at.config = nil
//	at.TriggeredAt = time.Now()
//}
