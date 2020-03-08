package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gitlab.com/zer0pts/zer0ptsctfd/scoreserver/model"
	"gitlab.com/zer0pts/zer0ptsctfd/scoreserver/service"

	"github.com/gorilla/websocket"
)

const (
	InvalidRequestMessage         = "invalid request"
	UnauthorizedMessage           = "unauthorized"
	RegisteredMessage             = "registered"
	LoginMessage                  = "logged in"
	LogoutMessage                 = "logged out"
	PasswordResetTokenSentMessage = "password reset token sent to your email"
	PasswordResetMessage          = "your password is updated"
	UpdateTeamNameMessage         = "teamname updated"
	UpdateCountryMessage          = "country updated"

	SubmissionLockMessage = "your team's submission is locked"

	ConfigUpdateMessage = "updated"

	ChallengeOpenMessage  = "OPEN: %s"
	ChallengeCloseMessage = "CLOSED: %s"

	WrongFlagMessage = "wrong flag"
	CorrectMessage   = "ALREADY SOLVED: %s"
	ValidMessage     = "SOLVED: %s"
)

var ()

type Server interface {
	Start(addr string) error
}

type server struct {
	app          service.App
	allowOrigins []string
	upgrader     websocket.Upgrader
}

func New(app service.App, allowOrigins []string) Server {
	return &server{
		app:          app,
		allowOrigins: allowOrigins,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				for _, o := range allowOrigins {
					if o == origin {
						return true
					}
				}
				return false
			},
		},
	}
}

func (s *server) Start(addr string) error {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     s.allowOrigins,
		AllowCredentials: true,
		AllowMethods:     []string{http.MethodGet, http.MethodPost},
	}))

	e.GET("/ws", s.wsHandler())
	e.GET("/user", s.userHandler(), s.loginMiddleware)
	e.GET("/team", s.teamHandler(), s.loginMiddleware)
	e.GET("/admin", s.adminHandler(), s.adminMiddleware)

	e.GET("/ctf", s.ctfHandler())

	e.POST("/login", s.loginHandler())
	e.POST("/logout", s.logoutHandler(), s.loginMiddleware)
	e.POST("/join-team", s.joinHandler())
	e.POST("/create-team", s.createHandler())

	e.POST("/reset-request", s.passwordResetRequestHandler())
	e.POST("/reset", s.passwordResetHandler())

	e.GET("/challenges", s.challengesHandler(), s.loginMiddleware, s.CTFStartedMiddleware)
	e.POST("/submit", s.submitHandler(), s.loginMiddleware, s.CTFStartedMiddleware)

	e.GET("/team/:id", s.teamPageHandler(), s.loginMiddleware)
	e.GET("/teams", s.teamsHandler())
	e.GET("/scorefeed", s.scoreFeedHandler())
	e.POST("/set-country", s.setCountryHandler(), s.loginMiddleware)
	e.POST("/set-teamname", s.setTeamNameHandler(), s.loginMiddleware)

	e.GET("/admin/challenges", s.adminChallengesHandler(), s.adminMiddleware)
	e.POST("/admin/set-challenges-status", s.adminSetChallengesStatusHandler(), s.adminMiddleware)
	e.POST("/admin/scoreupdate", s.adminScoreUpdateHandler(), s.adminMiddleware)
	e.POST("/set-ctf", s.setCTFHandler(), s.adminMiddleware)

	return e.Start(addr)
}

type LoginContext struct {
	echo.Context
	User *model.User
}

func (s *server) getLoginUser(c echo.Context) *model.User {
	cookie, err := c.Cookie(sessionKey)
	if err == nil && cookie.Value != "" {
		if user, _ := s.app.GetLoginUser(cookie.Value); user != nil {
			return user
		}
	}
	return nil
}

func (s *server) loginMiddleware(h echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := s.getLoginUser(c)
		if user == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"message": UnauthorizedMessage,
			})
		}
		return h(&LoginContext{c, user})
	}
}

func (s *server) adminMiddleware(h echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := s.getLoginUser(c)
		if user == nil || !user.IsAdmin {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"message": UnauthorizedMessage,
			})
		}
		return h(&LoginContext{c, user})
	}
}

func (s *server) CTFStartedMiddleware(h echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		t := time.Now()
		started, err := s.app.CTFStarted(t)
		if err != nil {
			return errorHandle(c, err)
		}
		if !started {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"message": service.CTFNotStartedYetMessage,
			})
		}
		return h(c)
	}
}

func errorHandle(c echo.Context, err error) error {
	if service.IsErrorMessage(err) {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": err.Error(),
		})
	}
	c.Logger().Error(err)
	return c.NoContent(http.StatusInternalServerError)
}

func (s *server) wsHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		user := s.getLoginUser(c)
		ws, err := s.upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return errorHandle(c, err)
		}

		client := s.app.NewClient(ws, user)
		s.app.Add(client)
		go client.SendHandler()
		return nil
	}
}

func (s *server) ctfHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		conf, err := s.app.GetConfig()
		if err != nil {
			return errorHandle(c, err)
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"config": conf,
		})
	}
}

func (s *server) adminHandler() echo.HandlerFunc {
	return func(cc echo.Context) error {
		c := cc.(*LoginContext)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"username": c.User.Username,
		})
	}
}

func (s *server) teamHandler() echo.HandlerFunc {
	return func(cc echo.Context) error {
		c := cc.(*LoginContext)
		team, err := s.app.GetUserTeam(c.User.ID)
		if err != nil {
			return errorHandle(c, err)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"team": team,
		})
	}
}

func (s *server) userHandler() echo.HandlerFunc {
	return func(cc echo.Context) error {
		c := cc.(*LoginContext)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"username": c.User.Username,
		})
	}
}

func (s *server) loginHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		req := new(struct {
			Username string `json:"username"`
			Password string `json:"Password"`
		})
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "invalid request",
			})
		}

		_, token, err := s.app.LoginUser(req.Username, req.Password)
		if err != nil {
			return errorHandle(c, err)
		}
		c.SetCookie(s.tokenCookie(token))

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": LoginMessage,
		})
	}
}

func (s *server) logoutHandler() echo.HandlerFunc {
	return func(cc echo.Context) error {
		c := cc.(*LoginContext)
		err := s.app.LogoutUser(c.User.ID)
		if err != nil {
			return errorHandle(c, err)
		}
		c.SetCookie(s.removeTokenCookie())
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": LogoutMessage,
		})
	}
}

func (s *server) joinHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		req := new(struct {
			Token    string `json:"token"`
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
		})
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": InvalidRequestMessage,
			})
		}
		err := s.app.JoinUserToTeam(req.Username, req.Email, req.Password, req.Token)
		if err != nil {
			return errorHandle(c, err)
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": RegisteredMessage,
		})
	}
}

func (s *server) createHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		req := new(struct {
			TeamName    string `json:"teamname"`
			CountryCode string `json:"country"`

			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
		})
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": InvalidRequestMessage,
			})
		}
		err := s.app.RegisterUserCreateTeam(req.Username, req.Email, req.Password, req.TeamName, req.CountryCode)
		if err != nil {
			return errorHandle(c, err)
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": RegisteredMessage,
		})
	}
}

func (s *server) passwordResetRequestHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		req := new(struct {
			Email string `json:"email"`
		})
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": InvalidRequestMessage,
			})
		}

		err := s.app.IssuePasswordResetToken(req.Email)
		if err != nil {
			return errorHandle(c, err)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": PasswordResetTokenSentMessage,
		})
	}
}

func (s *server) passwordResetHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		req := new(struct {
			Token    string `json:"token"`
			Password string `json:"password"`
		})
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": InvalidRequestMessage,
			})
		}

		err := s.app.ResetPassword(req.Token, req.Password)
		if err != nil {
			return errorHandle(c, err)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": PasswordResetMessage,
		})
	}
}

func (s *server) challengesHandler() echo.HandlerFunc {
	return func(cc echo.Context) error {
		c := cc.(*LoginContext)

		chals, err := s.app.ListOpenChallenges()
		if err != nil {
			return errorHandle(c, err)
		}
		userchals := make([]*model.UserChallengeInfo, 0, len(chals))
		for _, chal := range chals {
			uc, err := model.UserChallenge(chal)
			if err != nil {
				c.Logger().Error(err)
				continue
			}
			userchals = append(userchals, uc)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"challenges": userchals,
		})
	}
}

func (s *server) submitHandler() echo.HandlerFunc {
	return func(cc echo.Context) error {
		c := cc.(*LoginContext)

		req := new(struct {
			Flag string `json:"flag"`
		})
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": InvalidRequestMessage,
			})
		}

		t, err := s.app.GetUserTeam(c.User.ID)
		if err != nil {
			return errorHandle(c, err)
		}

		submittable, err := s.app.CheckSubmittable(t.ID)
		if err != nil {
			return errorHandle(c, err)
		}
		if !submittable {
			return c.JSON(http.StatusForbidden, map[string]interface{}{
				"message": SubmissionLockMessage,
			})
		}

		conf, err := s.app.GetConfig()
		if err != nil {
			return errorHandle(c, err)
		}

		chal, valid, err := s.app.SubmitFlag(c.User, req.Flag)
		if err != nil {
			return errorHandle(c, err)
		}

		// if wrong flag
		if chal == nil {
			cnt, err := s.app.AddWrongCount(t.ID, time.Duration(conf.LockSecond)*time.Second)
			if err != nil {
				return errorHandle(c, err)
			}
			if cnt >= conf.LockCount {
				err = s.app.LockSubmission(t.ID, time.Duration(conf.LockDuration)*time.Second)
				if err != nil {
					return errorHandle(c, err)
				}
			}

			return c.JSON(http.StatusOK, map[string]interface{}{
				"message": WrongFlagMessage,
			})
		}

		// if correct/valid flag
		if valid {
			if chal.IsDynamic {
				err = s.app.RecalcScore(conf.MinScore, chal.BaseScore, conf.EasySolves, conf.MediumSolves, chal.ID)
				if err != nil {
					c.Logger().Error(err)
				}
			}
			select {
			default:
				chal, err = s.app.GetChallenge(chal.ID)
				if err != nil {
					c.Logger().Error(err)
					break
				}
				userchal, err := model.UserChallenge(chal)
				if err != nil {
					c.Logger().Error(err)
					break
				}
				s.wsChallengeUpdate(userchal)
			}

			return c.JSON(http.StatusOK, map[string]interface{}{
				"message": fmt.Sprintf(ValidMessage, chal.Name),
			})
		} else {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"message": fmt.Sprintf(CorrectMessage, chal.Name),
			})
		}
		return nil
	}
}

func (s *server) teamPageHandler() echo.HandlerFunc {
	return func(cc echo.Context) error {
		c := cc.(*LoginContext)
		tidStr := c.Param("id")
		tid, err := strconv.Atoi(tidStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": InvalidRequestMessage,
			})
		}
		t, err := s.app.GetTeam(uint32(tid))
		if err != nil {
			return errorHandle(c, err)
		}

		if t.ID != c.User.TeamID {
			t.Token = ""
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"team": t,
		})
	}
}

func (s *server) teamsHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		teams, err := s.app.GetTeams()
		if err != nil {
			return errorHandle(c, err)
		}

		for i := 0; i < len(teams); i++ {
			teams[i].Token = ""
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"teams": teams,
		})
	}
}

func (s *server) scoreFeedHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		chals, err := s.app.ListOpenChallenges()
		if err != nil {
			return errorHandle(c, err)
		}
		chalMap := make(map[uint32]*model.Challenge)
		for i := 0; i < len(chals); i++ {
			chalMap[chals[i].ID] = chals[i]
		}

		teams, err := s.app.GetTeams()
		if err != nil {
			return errorHandle(c, err)
		}

		xs := make([]struct {
			Pos            int    `json:"pos"`
			Team           string `json:"team"`
			Score          int    `json:"score"`
			LastSubmission int64  `json:"-"`
		}, len(teams))
		for i := 0; i < len(teams); i++ {
			score := 0
			var lastSub int64 = 0
			for _, s := range teams[i].Submissions {
				if s.ChallengeID == nil {
					continue
				}
				cid := *s.ChallengeID
				if !chalMap[cid].IsQuestionary && lastSub < s.SubmittedAt {
					lastSub = s.SubmittedAt
				}
				score += chalMap[cid].Score
			}
			xs[i].Score = score
			xs[i].LastSubmission = lastSub
			xs[i].Team = teams[i].Teamname
		}

		sort.Slice(xs, func(i, j int) bool {
			if xs[i].Score == xs[j].Score {
				return xs[i].LastSubmission < xs[j].LastSubmission
			}
			return xs[i].Score > xs[j].Score
		})

		for i := 0; i < len(xs); i++ {
			xs[i].Pos = i + 1
			if i != 0 && xs[i].Score == xs[i-1].Score && xs[i].LastSubmission == xs[i-1].LastSubmission {
				xs[i].Pos = xs[i-1].Pos
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"standings": xs,
		})
	}
}

func (s *server) setTeamNameHandler() echo.HandlerFunc {
	return func(cc echo.Context) error {
		c := cc.(*LoginContext)
		req := new(struct {
			Teamame string `json:"teamname"`
		})
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": InvalidRequestMessage,
			})
		}
		team, err := s.app.GetUserTeam(c.User.ID)
		if err != nil {
			return errorHandle(c, err)
		}
		err = s.app.UpdateTeamName(team.ID, req.Teamame)
		if err != nil {
			return errorHandle(c, err)
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": UpdateTeamNameMessage,
		})
	}
}

func (s *server) setCountryHandler() echo.HandlerFunc {
	return func(cc echo.Context) error {
		c := cc.(*LoginContext)
		req := new(struct {
			Country string `json:"country"`
		})
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": InvalidRequestMessage,
			})
		}
		team, err := s.app.GetUserTeam(c.User.ID)
		if err != nil {
			return errorHandle(c, err)
		}
		err = s.app.UpdateTeamCountry(team.ID, req.Country)
		if err != nil {
			return errorHandle(c, err)
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": UpdateCountryMessage,
		})
	}
}

func (s *server) setCTFHandler() echo.HandlerFunc {
	return func(cc echo.Context) error {
		req := new(struct {
			CTFName      string `json:"ctf_name"`
			StartAt      int64  `json:"start_at"`
			EndAt        int64  `json:"end_at"`
			LockSecond   int    `json:"lock_second"`
			LockCount    int    `json:"lock_count"`
			LockDuration int    `json:"lock_duration"`
			EasySolves   int    `json:"easy_solves"`
			MediumSolves int    `json:"medium_solves"`
			MinScore     int    `json:"min_score"`
		})
		if err := cc.Bind(req); err != nil {
			return cc.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": InvalidRequestMessage,
			})
		}

		if err := s.app.SetCTFName(req.CTFName); err != nil {
			return errorHandle(cc, err)
		}
		if err := s.app.SetStartAt(req.StartAt); err != nil {
			return errorHandle(cc, err)
		}
		if err := s.app.SetEndAt(req.EndAt); err != nil {
			return errorHandle(cc, err)
		}
		if err := s.app.SetLock(req.LockSecond, req.LockCount, req.LockDuration); err != nil {
			return errorHandle(cc, err)
		}
		if err := s.app.SetSolves(req.EasySolves, req.MediumSolves); err != nil {
			return errorHandle(cc, err)
		}
		if err := s.app.SetMinScore(req.MinScore); err != nil {
			return errorHandle(cc, err)
		}

		return cc.JSON(http.StatusOK, map[string]interface{}{
			"message": ConfigUpdateMessage,
		})
	}
}

func (s *server) adminChallengesHandler() echo.HandlerFunc {
	return func(cc echo.Context) error {
		chals, err := s.app.ListAllChallenges()
		if err != nil {
			return errorHandle(cc, err)
		}
		return cc.JSON(http.StatusOK, map[string]interface{}{
			"challenges": chals,
		})
	}
}

func (s *server) adminSetChallengesStatusHandler() echo.HandlerFunc {
	return func(cc echo.Context) error {
		req := new(struct {
			Challenges []struct {
				ID     uint32 `json:"id"`
				IsOpen bool   `json:"is_open"`
			} `json:"challenges"`
		})
		if err := cc.Bind(req); err != nil {
			return cc.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": InvalidRequestMessage,
			})
		}
		chals, err := s.app.ListAllChallenges()
		if err != nil {
			return errorHandle(cc, err)
		}

		for _, c := range req.Challenges {
			i := 0
			for ; i < len(chals); i++ {
				if c.ID == chals[i].ID {
					break
				}
			}
			if i == len(chals) {
				continue
			}
			if c.IsOpen == chals[i].IsOpen {
				continue
			}

			if c.IsOpen {
				if err := s.app.OpenChallenge(c.ID); err == nil {
					s.wsMessage(fmt.Sprintf(ChallengeOpenMessage, chals[i].Name))
					uchal, err := model.UserChallenge(chals[i])
					if err == nil {
						s.wsChallengeUpdate(uchal)
					} else {
						cc.Logger().Error(err)
					}
				} else {
					cc.Logger().Error(err)
				}
			} else {
				if err := s.app.CloseChallenge(c.ID); err == nil {
					s.wsMessage(fmt.Sprintf(ChallengeCloseMessage, chals[i].Name))
					s.wsChallengeClose(c.ID)
				} else {
					cc.Logger().Error(err)
				}
			}
		}
		return cc.NoContent(http.StatusOK)
	}
}

func (s *server) adminScoreUpdateHandler() echo.HandlerFunc {
	return func(cc echo.Context) error {
		chals, err := s.app.ListOpenChallenges()
		if err != nil {
			return errorHandle(cc, err)
		}

		conf, err := s.app.GetConfig()
		if err != nil {
			return errorHandle(cc, err)
		}

		for _, chal := range chals {
			if !chal.IsDynamic {
				continue
			}
			err = s.app.RecalcScore(conf.MinScore, chal.BaseScore, conf.EasySolves, conf.MediumSolves, chal.ID)
			if err != nil {
				return errorHandle(cc, err)
			}
		}
		return cc.NoContent(http.StatusOK)
	}
}

const sessionKey = "zer0ptsctf"

func (s *server) tokenCookie(token string) *http.Cookie {
	return &http.Cookie{
		Name:     sessionKey,
		Value:    token,
		Expires:  time.Now().Add(service.TokenLimit),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}
func (s *server) removeTokenCookie() *http.Cookie {
	return &http.Cookie{
		Name:     sessionKey,
		Value:    "",
		Expires:  time.Time{},
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}

func (s *server) wsMessage(msg string) error {
	data, err := json.Marshal(struct {
		Type    string `json:"type"`
		Message string `json:"value"`
	}{
		Type:    "message",
		Message: msg,
	})
	if err != nil {
		return err
	}

	s.app.Send(data, true, false)
	return nil
}

func (s *server) wsChallengeClose(cid uint32) error {
	data, err := json.Marshal(struct {
		Type        string `json:"type"`
		ChallengeID uint32 `json:"value"`
	}{
		Type:        "challengeClose",
		ChallengeID: cid,
	})
	if err != nil {
		return err
	}

	s.app.Send(data, true, false)
	return nil

}

func (s *server) wsChallengeUpdate(chal *model.UserChallengeInfo) error {
	data, err := json.Marshal(struct {
		Type      string                  `json:"type"`
		Challenge model.UserChallengeInfo `json:"value"`
	}{
		Type:      "challengeUpdate",
		Challenge: *chal,
	})
	if err != nil {
		return err
	}

	s.app.Send(data, true, false)
	return nil
}
