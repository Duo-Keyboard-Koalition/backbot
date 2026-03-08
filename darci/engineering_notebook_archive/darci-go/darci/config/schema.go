package config

// Config holds the complete darci configuration.
type Config struct {
	Providers ProvidersConfig `json:"providers"`
	Channels  ChannelsConfig  `json:"channels"`
	Tools     ToolsConfig     `json:"tools"`
}

// ProvidersConfig holds LLM provider configurations.
type ProvidersConfig struct {
	Gemini *GeminiConfig `json:"gemini,omitempty"`
}

// GeminiConfig holds Gemini provider configuration.
type GeminiConfig struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url,omitempty"`
	Model   string `json:"model,omitempty"`
}

// ChannelsConfig holds all channel configurations.
type ChannelsConfig struct {
	Telegram  *TelegramChannelConfig  `json:"telegram,omitempty"`
	Discord   *DiscordChannelConfig   `json:"discord,omitempty"`
	WhatsApp  *WhatsAppChannelConfig  `json:"whatsapp,omitempty"`
	Slack     *SlackChannelConfig     `json:"slack,omitempty"`
	Email     *EmailChannelConfig     `json:"email,omitempty"`
	Feishu    *FeishuChannelConfig    `json:"feishu,omitempty"`
	QQ        *QQChannelConfig        `json:"qq,omitempty"`
	DingTalk  *DingTalkChannelConfig  `json:"dingtalk,omitempty"`
	Mochat    *MochatChannelConfig    `json:"mochat,omitempty"`
	Matrix    *MatrixChannelConfig    `json:"matrix,omitempty"`
}

// TelegramChannelConfig holds Telegram channel configuration.
type TelegramChannelConfig struct {
	Enabled        bool        `json:"enabled"`
	Token          string      `json:"token"`
	AllowFrom      []string    `json:"allow_from"`
	Proxy          string      `json:"proxy,omitempty"`
	ReplyToMessage bool        `json:"reply_to_message"`
	ReactEmoji     string      `json:"react_emoji"`
	Voice          VoiceConfig `json:"voice"`
}

// VoiceConfig holds voice/TTS configuration.
type VoiceConfig struct {
	Enabled bool   `json:"enabled"`
	Voice   string `json:"voice"`
	Always  bool   `json:"always"`
}

// DiscordChannelConfig holds Discord channel configuration.
type DiscordChannelConfig struct {
	Enabled   bool     `json:"enabled"`
	Token     string   `json:"token"`
	AllowFrom []string `json:"allow_from"`
}

// WhatsAppChannelConfig holds WhatsApp channel configuration.
type WhatsAppChannelConfig struct {
	Enabled   bool     `json:"enabled"`
	AllowFrom []string `json:"allow_from"`
}

// SlackChannelConfig holds Slack channel configuration.
type SlackChannelConfig struct {
	Enabled        bool            `json:"enabled"`
	Mode           string          `json:"mode"`
	BotToken       string          `json:"bot_token"`
	AppToken       string          `json:"app_token"`
	GroupPolicy    string          `json:"group_policy"`
	GroupAllowFrom []string        `json:"group_allow_from"`
	DM             SlackDMConfig   `json:"dm"`
	ReactEmoji     string          `json:"react_emoji"`
	ReplyInThread  bool            `json:"reply_in_thread"`
}

// SlackDMConfig holds Slack DM configuration.
type SlackDMConfig struct {
	Enabled   bool     `json:"enabled"`
	Policy    string   `json:"policy"`
	AllowFrom []string `json:"allow_from"`
}

// EmailChannelConfig holds Email channel configuration.
type EmailChannelConfig struct {
	Enabled          bool   `json:"enabled"`
	ConsentGranted   bool   `json:"consent_granted"`
	IMAPHost         string `json:"imap_host"`
	IMAPPort         int    `json:"imap_port"`
	IMAPUsername     string `json:"imap_username"`
	IMAPPassword     string `json:"imap_password"`
	IMAPMailbox      string `json:"imap_mailbox"`
	IMAPUseSSL       bool   `json:"imap_use_ssl"`
	SMTPHost         string `json:"smtp_host"`
	SMTPPort         int    `json:"smtp_port"`
	SMTPUsername     string `json:"smtp_username"`
	SMTPPassword     string `json:"smtp_password"`
	SMTPUseTLS       bool   `json:"smtp_use_tls"`
	SMTPUseSSL       bool   `json:"smtp_use_ssl"`
	FromAddress      string `json:"from_address"`
	AutoReplyEnabled bool   `json:"auto_reply_enabled"`
	PollIntervalSecs int    `json:"poll_interval_seconds"`
	MarkSeen         bool   `json:"mark_seen"`
	MaxBodyChars     int    `json:"max_body_chars"`
	SubjectPrefix    string `json:"subject_prefix"`
	AllowFrom        []string `json:"allow_from"`
}

// FeishuChannelConfig holds Feishu channel configuration.
type FeishuChannelConfig struct {
	Enabled           bool     `json:"enabled"`
	AppID             string   `json:"app_id"`
	AppSecret         string   `json:"app_secret"`
	EncryptKey        string   `json:"encrypt_key,omitempty"`
	VerificationToken string   `json:"verification_token,omitempty"`
	AllowFrom         []string `json:"allow_from"`
}

// QQChannelConfig holds QQ channel configuration.
type QQChannelConfig struct {
	Enabled   bool     `json:"enabled"`
	AppID     string   `json:"app_id"`
	Secret    string   `json:"secret"`
	AllowFrom []string `json:"allow_from"`
}

// DingTalkChannelConfig holds DingTalk channel configuration.
type DingTalkChannelConfig struct {
	Enabled     bool     `json:"enabled"`
	ClientID    string   `json:"client_id"`
	ClientSecret string  `json:"client_secret"`
	AllowFrom   []string `json:"allow_from"`
}

// MochatChannelConfig holds Mochat channel configuration.
type MochatChannelConfig struct {
	Enabled      bool     `json:"enabled"`
	BaseURL      string   `json:"base_url"`
	SocketURL    string   `json:"socket_url"`
	SocketPath   string   `json:"socket_path"`
	ClawToken    string   `json:"claw_token"`
	AgentUserID  string   `json:"agent_user_id"`
	Sessions     []string `json:"sessions"`
	Panels       []string `json:"panels"`
	ReplyDelayMode string `json:"reply_delay_mode"`
	ReplyDelayMs   int    `json:"reply_delay_ms"`
}

// MatrixChannelConfig holds Matrix channel configuration.
type MatrixChannelConfig struct {
	Enabled         bool     `json:"enabled"`
	Homeserver      string   `json:"homeserver"`
	UserID          string   `json:"user_id"`
	AccessToken     string   `json:"access_token"`
	DeviceID        string   `json:"device_id"`
	E2EEEnabled     bool     `json:"e2ee_enabled"`
	AllowFrom       []string `json:"allow_from"`
	GroupPolicy     string   `json:"group_policy"`
	GroupAllowFrom  []string `json:"group_allow_from"`
	AllowRoomMentions bool   `json:"allow_room_mentions"`
	MaxMediaBytes   int      `json:"max_media_bytes"`
}

// ToolsConfig holds tool configurations.
type ToolsConfig struct {
	MCPServers        map[string]MCPServerConfig `json:"mcp_servers"`
	RestrictToWorkspace bool                      `json:"restrict_to_workspace"`
	ExecPathAppend    string                     `json:"exec_path_append"`
	ToolTimeout       int                        `json:"tool_timeout"`
}

// MCPServerConfig holds MCP server configuration.
type MCPServerConfig struct {
	Command     string            `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	URL         string            `json:"url,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	ToolTimeout int               `json:"tool_timeout,omitempty"`
}

// SecurityConfig holds security configuration.
type SecurityConfig struct {
	RestrictToWorkspace bool   `json:"restrict_to_workspace"`
	ExecPathAppend      string `json:"exec_path_append"`
}
