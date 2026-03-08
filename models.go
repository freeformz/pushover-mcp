package main

// MessageRequest represents a request to send a Pushover message.
type MessageRequest struct {
	Token     string `json:"token"`
	User      string `json:"user"`
	Message   string `json:"message"`
	Title     string `json:"title,omitempty"`
	Priority  int    `json:"priority,omitempty"`
	Sound     string `json:"sound,omitempty"`
	Device    string `json:"device,omitempty"`
	URL       string `json:"url,omitempty"`
	URLTitle  string `json:"url_title,omitempty"`
	HTML      int    `json:"html,omitempty"`
	Monospace int    `json:"monospace,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
	TTL       int    `json:"ttl,omitempty"`
	Retry     int    `json:"retry,omitempty"`
	Expire    int    `json:"expire,omitempty"`
	Callback  string `json:"callback,omitempty"`
	Tags      string `json:"tags,omitempty"`
}

// MessageResponse represents a response from the Pushover messages API.
type MessageResponse struct {
	Status  int      `json:"status"`
	Request string   `json:"request"`
	Receipt string   `json:"receipt,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}

// ReceiptResponse represents a response from the Pushover receipt polling API.
type ReceiptResponse struct {
	Status               int      `json:"status"`
	Request              string   `json:"request"`
	Acknowledged         int      `json:"acknowledged"`
	AcknowledgedAt       int64    `json:"acknowledged_at"`
	AcknowledgedBy       string   `json:"acknowledged_by"`
	AcknowledgedByDevice string   `json:"acknowledged_by_device"`
	LastDeliveredAt      int64    `json:"last_delivered_at"`
	Expired              int      `json:"expired"`
	ExpiresAt            int64    `json:"expires_at"`
	CalledBack           int      `json:"called_back"`
	CalledBackAt         int64    `json:"called_back_at"`
	Errors               []string `json:"errors,omitempty"`
}

// CancelResponse represents a response from the Pushover cancel API.
type CancelResponse struct {
	Status  int      `json:"status"`
	Request string   `json:"request"`
	Errors  []string `json:"errors,omitempty"`
}
