package shared

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
)

type flashPayload struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

type hxTriggerPayload struct {
	ShowFlash flashPayload `json:"showFlash"`
}

// SetFlash sets the HX-Trigger header with a success flash message.
func SetFlash(c *gin.Context, message string) {
	setFlashHeader(c, message, "success")
}

// SetFlashError sets the HX-Trigger header with an error flash message.
func SetFlashError(c *gin.Context, message string) {
	setFlashHeader(c, message, "error")
}

func setFlashHeader(c *gin.Context, message, flashType string) {
	payload := hxTriggerPayload{
		ShowFlash: flashPayload{
			Message: message,
			Type:    flashType,
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	c.Header("HX-Trigger", string(b))
}
