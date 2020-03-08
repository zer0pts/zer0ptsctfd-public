package model

type Token struct {
	UserID    uint32 `db:"user_id"`
	Token     string `db:"token"`
	ExpiresAt int64  `db:"expires_at"`

	CreatedAt string `db:"created_at" json:"-"`
	UpdatedAt string `db:"updated_at" json:"-"`
}

type PasswordResetToken struct {
	UserID    uint32 `db:"user_id"`
	Token     string `db:"token"`
	ExpiresAt int64  `db:"expires_at"`

	CreatedAt string `db:"created_at" json:"-"`
	UpdatedAt string `db:"updated_at" json:"-"`
}

type User struct {
	ID           uint32 `db:"id" json:"id"`
	Username     string `db:"username" json:"username"`
	Email        string `db:"email" json:"-"`
	PasswordHash string `db:"password_hash" json:"-"`
	Password     string
	IconPath     *string `db:"icon_path" json:"icon_path"`
	TeamID       uint32  `db:"team_id" json:"team_id"`
	IsHidden     bool    `db:"is_hidden" json:"-"`
	IsAdmin      bool    `db:"is_admin" json:"-"`

	CreatedAt string `db:"created_at" json:"-"`
	UpdatedAt string `db:"updated_at" json:"-"`
}

type Team struct {
	ID          uint32 `db:"id" json:"id"`
	Teamname    string `db:"teamname" json:"teamname"`
	Token       string `db:"token" json:"token"`
	CountryCode string `db:"country_code" json:"country_code"`
	IsHidden    bool   `db:"is_hidden" json:"-"`

	Submissions []*Submission `json:"submissions"`
	Users       []*User       `json:"users"`

	CreatedAt string `db:"created_at" json:"-"`
	UpdatedAt string `db:"updated_at" json:"-"`
}

type Config struct {
	CTFName string `db:"ctf_name" json:"ctf_name"`
	StartAt int64  `db:"start_at" json:"start_at"`
	EndAt   int64  `db:"end_at" json:"end_at"`

	LockCount    int `db:"lock_count" json:"lock_count"`
	LockSecond   int `db:"lock_second" json:"lock_second"`
	LockDuration int `db:"lock_duration" json:"lock_duration"`

	MinScore     int `db:"min_score" json:"min_score"`
	EasySolves   int `db:"easy_solves" json:"easy_solves"`
	MediumSolves int `db:"medium_solves" json:"medium_solves"`

	CreatedAt string `db:"created_at" json:"-"`
	UpdatedAt string `db:"updated_at" json:"-"`
}

type Challenge struct {
	ID            uint32   `db:"id" json:"id"`
	Name          string   `db:"name" json:"name"`
	Flag          string   `db:"flag" json:"flag"`
	Description   string   `db:"description" json:"description"`
	Category      string   `db:"category" json:"category"`
	Difficulty    string   `db:"difficulty" json:"difficulty"`
	Tags          []string `db:"tags" json:"tags"`
	Attachments   []string `db:"attachments" json:"attachments"`
	Author        string   `db:"author" json:"author"`
	BaseScore     int      `db:"base_score" json:"base_score"`
	Score         int      `json:"score"`
	IsOpen        bool     `db:"is_open" json:"is_open"`
	IsDynamic     bool     `db:"is_dynamic" json:"is_dynamic"`
	IsQuestionary bool     `db:"is_questionary" json:"is_questionary"`
	Host          *string  `db:"host" json:"host"`
	Port          *string  `db:"port" json:"port"`
	SolveTeams    []uint32 `json:"solveteams"`

	CreatedAt string `db:"created_at" json:"-"`
	UpdatedAt string `db:"updated_at" json:"-"`
}

type UserChallengeInfo struct {
	ID            uint32   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	Difficulty    string   `json:"difficulty"`
	Tags          []string `json:"tags"`
	Attachments   []string `json:"attachments"`
	Author        string   `json:"author"`
	Score         int      `json:"score"`
	SolveTeams    []uint32 `json:"solveteams"`
	IsQuestionary bool     `json:"is_questionary"`

	CreatedAt string `db:"created_at" json:"-"`
	UpdatedAt string `db:"updated_at" json:"-"`
}
type Tag struct {
	ID          uint32 `db:"id"`
	ChallengeID uint32 `db:"challenge_id"`
	Tag         string `db:"tag"`

	CreatedAt string `db:"created_at" json:"-"`
	UpdatedAt string `db:"updated_at" json:"-"`
}
type Attachment struct {
	ID          uint32 `db:"id"`
	ChallengeID uint32 `db:"challenge_id"`
	URL         string `db:"url"`

	CreatedAt string `db:"created_at" json:"-"`
	UpdatedAt string `db:"updated_at" json:"-"`
}
type Submission struct {
	ID          uint32  `db:"id" json:"id"`
	ChallengeID *uint32 `db:"challenge_id" json:"challenge_id"`
	UserID      *uint32 `db:"user_id" json:"user_id"`
	TeamID      *uint32 `db:"team_id" json:"team_id"`
	Flag        string  `db:"flag" json:"-"`
	SubmittedAt int64   `db:"submitted_at" json:"submitted_at"`

	IsCorrect bool `db:"is_correct" json:"-"`
	IsValid   bool `db:"is_valid" json:"-"`

	CreatedAt string `db:"created_at" json:"-"`
	UpdatedAt string `db:"updated_at" json:"-"`
}
