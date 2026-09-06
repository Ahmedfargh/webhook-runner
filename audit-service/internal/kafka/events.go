package kafka

type AuditEvent struct {
	ID           string `json:"id"`
	ActorID      string `json:"actor_id"`
	ActorType    string `json:"actor_type"` // USER, ADMIN, SYSTEM, SERVICE
	ActorName    string `json:"actor_name"`
	ActorEmail   string `json:"actor_email"`
	ServiceName  string `json:"service_name"`
	Action       string `json:"action"`   // CREATE, UPDATE, DELETE, LOGIN, OVERRIDE, ROTATE_SECRET, DISPATCH
	Resource     string `json:"resource"` // USER, ADMIN, ROLE, PERMISSION, PLAN, SUBSCRIPTION, INVOICE, APP, WEBHOOK
	ResourceID   string `json:"resource_id"`
	BeforeJSON   string `json:"before_json"`
	AfterJSON    string `json:"after_json"`
	IPAddress    string `json:"ip_address"`
	UserAgent    string `json:"user_agent"`
	Status       string `json:"status"` // SUCCESS, FAILED
	ErrorMessage string `json:"error_message"`
	Timestamp    int64  `json:"timestamp"`
}
