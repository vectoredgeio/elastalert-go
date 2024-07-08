package alerts

import "log"

type Alert interface {
	Send(message string) error
}

func SendAlerts(alerts []Alert, message string) {
	for _, alert := range alerts {
		err := alert.Send(message)
		if err != nil {
			log.Printf("Error sending alert: %v", err)
		}
	}
}
