package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/cristalhq/jwt"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

const (
	postMeetingKey = "post_meeting_"
	commandTrigger = "jitsi"

	botName        = "Jitsi"
	botDescription = "Created by the Jitsi Meet plugin."
)

type Plugin struct {
	plugin.MattermostPlugin

	// botUserID of the created bot account.
	botUserID string

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

func (p *Plugin) OnActivate() error {
	config := p.getConfiguration()
	if err := config.IsValid(); err != nil {
		return err
	}

	botUserID, err := p.Helpers.EnsureBot(&model.Bot{
		Username:    botName,
		DisplayName: botName,
		Description: botDescription,
	})
	if err != nil {
		return errors.Wrap(err, "failed to ensure bot account")
	}
	p.botUserID = botUserID

	bundlePath, err := p.API.GetBundlePath()
	if err != nil {
		return errors.Wrap(err, "couldn't get bundle path")
	}

	botProfileImage, err := ioutil.ReadFile(filepath.Join(bundlePath, "public", "jitsi_logo.png"))
	if err != nil {
		return errors.Wrap(err, "couldn't read profile image")
	}

	if appErr := p.API.SetProfileImage(botUserID, botProfileImage); appErr != nil {
		return errors.Wrap(appErr, "couldn't set profile image")
	}

	botIconImage, err := ioutil.ReadFile(filepath.Join(bundlePath, "assets", "jitsi_logo.svg"))
	if err != nil {
		return errors.Wrap(err, "couldn't read profile image")
	}

	if appErr := p.API.SetBotIconImage(botUserID, botIconImage); appErr != nil {
		return errors.Wrap(appErr, "couldn't set icon image")
	}

	if err := p.API.RegisterCommand(&model.Command{
		Trigger:          commandTrigger,
		DisplayName:      botName,
		Description:      "Jitsi Meet slash command integration.",
		AutoComplete:     true,
		AutoCompleteHint: "[roomname]",
		AutoCompleteDesc: "Create a Jitsi Meeting",
	}); err != nil {
		return errors.Wrapf(err, "failed to register %s command", "jitsi")
	}

	return nil
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	switch path := r.URL.Path; path {
	case "/api/v1/meetings":
		p.handleStartMeeting(w, r)
	default:
		http.NotFound(w, r)
	}
}

type StartMeetingRequest struct {
	ChannelId string `json:"channel_id"`
	Personal  bool   `json:"personal"`
	Topic     string `json:"topic"`
	MeetingId int    `json:"meeting_id"`
}

// Claims extents cristalhq/jwt standard claims to add jitsi-web-token specific fields
type Claims struct {
	jwt.StandardClaims
	Room string `json:"room,omitempty"`
}

// MarshalBinary default marshaling to JSON.
func (c Claims) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

func encodeJitsiMeetingID(meeting string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	return reg.ReplaceAllString(meeting, "")
}

func (p *Plugin) generateMeetingID(topic string) string {
	var meetingID string
	if len(topic) < 1 {
		meetingID = generateRoomWithoutSeparator()
	} else {
		meetingID = encodeJitsiMeetingID(topic)
	}

	return meetingID
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	trigger := strings.TrimPrefix(strings.Fields(args.Command)[0], "/")
	switch trigger {
	case commandTrigger:
		return p.startMeetingCommand(args), nil

	default:
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         fmt.Sprintf("Unknown command: " + args.Command),
		}, nil
	}
}

func (p *Plugin) startMeetingCommand(args *model.CommandArgs) *model.CommandResponse {
	channel, _ := p.API.GetChannel(args.ChannelId)
	team, _ := p.API.GetTeam(args.TeamId)
	topic := fmt.Sprintf("%s_%s", team.Name, channel.Name)
	if len(strings.Fields(args.Command)) > 1 {
		topic = strings.Fields(args.Command)[1]
	}

	meetingLinkValidUntil := time.Now().Add(time.Duration(p.getConfiguration().JitsiLinkValidTime) * time.Minute)

	p.makeJitsiPost(args, p.generateMeetingID(topic), false, meetingLinkValidUntil)
	return &model.CommandResponse{}
}

func (p *Plugin) handleStartMeeting(w http.ResponseWriter, r *http.Request) {
	if err := p.getConfiguration().IsValid(); err != nil {
		http.Error(w, err.Error(), http.StatusTeapot)
		return
	}

	userId := r.Header.Get("Mattermost-User-Id")

	if userId == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	var req StartMeetingRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var user *model.User
	var err *model.AppError
	user, err = p.API.GetUser(userId)
	if err != nil {
		http.Error(w, err.Error(), err.StatusCode)
	}

	if _, err = p.API.GetChannelMember(req.ChannelId, user.Id); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var meetingID string
	meetingID = encodeJitsiMeetingID(req.Topic)
	if len(req.Topic) < 1 {
		meetingID = generateRoomWithoutSeparator()
	}
	jitsiURL := strings.TrimSpace(p.getConfiguration().JitsiURL)
	jitsiURL = strings.TrimRight(jitsiURL, "/")
	meetingURL := jitsiURL + "/" + meetingID

	var meetingLinkValidUntil = time.Time{}
	JWTMeeting := p.getConfiguration().JitsiJWT

	if JWTMeeting {
		signer, err2 := jwt.NewHS256([]byte(p.getConfiguration().JitsiAppSecret))
		if err2 != nil {
			log.Printf("Error generating new HS256 signer: %v", err2)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		builder := jwt.NewTokenBuilder(signer)

		// Error check is done in configuration.IsValid()
		jURL, _ := url.Parse(p.getConfiguration().JitsiURL)

		meetingLinkValidUntil = time.Now().Add(time.Duration(p.getConfiguration().JitsiLinkValidTime) * time.Minute)

		claims := Claims{}
		claims.Issuer = p.getConfiguration().JitsiAppID
		claims.Audience = []string{p.getConfiguration().JitsiAppID}
		claims.ExpiresAt = jwt.Timestamp(meetingLinkValidUntil.Unix())
		claims.Subject = jURL.Hostname()
		claims.Room = meetingID

		token, err2 := builder.Build(claims)
		if err2 != nil {
			log.Printf("Error building JWT: %v", err2)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		meetingURL = meetingURL + "?jwt=" + string(token.Raw())
	}

	post := &model.Post{
		UserId:    user.Id,
		ChannelId: req.ChannelId,
		Message:   fmt.Sprintf("Meeting started at %s.", meetingURL),
		Type:      "custom_jitsi",
		Props: map[string]interface{}{
			"meeting_id":              meetingID,
			"meeting_link":            meetingURL,
			"jwt_meeting":             JWTMeeting,
			"jwt_meeting_valid_until": meetingLinkValidUntil.Format("2006-01-02 15:04:05 Z07:00"),
			"meeting_personal":        false,
			"meeting_topic":           req.Topic,
			"from_webhook":            "true",
			"override_username":       "Jitsi",
			"override_icon_url":       "https://s3.amazonaws.com/mattermost-plugin-media/Zoom+App.png",
		},
	}

	if _, err = p.API.CreatePost(post); err != nil {
		http.Error(w, err.Error(), err.StatusCode)
		return
	}

	err = p.API.KVSet(fmt.Sprintf("%v%v", postMeetingKey, meetingID), []byte(post.Id))
	if err != nil {
		http.Error(w, err.Error(), err.StatusCode)
		return
	}

	w.Write([]byte(fmt.Sprintf("%v", meetingID)))
}

func (p *Plugin) makeJitsiPost(args *model.CommandArgs, meetingID string, JWTMeeting bool, meetingExpiry time.Time) {
	jitsiURL := strings.TrimRight(strings.TrimSpace(p.getConfiguration().JitsiURL), "/")
	if len(jitsiURL) == 0 {
		jitsiURL = "https://meet.jit.si"
	}
	meetingURL := jitsiURL + "/" + meetingID

	post := &model.Post{
		UserId:    p.botUserID,
		ChannelId: args.ChannelId,
		Message:   fmt.Sprintf("Meeting started at %s.", meetingURL),
		Type:      "custom_jitsi",
		Props: map[string]interface{}{
			"meeting_id":              meetingID,
			"meeting_link":            meetingURL,
			"jwt_meeting":             JWTMeeting,
			"jwt_meeting_valid_until": meetingExpiry.Format("2006-01-02 15:04:05 Z07:00"),
			"meeting_personal":        false,
			"meeting_topic":           "Jitsi Meeting",
			"from_webhook":            "true",
			"override_username":       "Jitsi",
			"override_icon_url":       "/plugins/jitsi/public/jitsi_logo.png",
		},
	}

	var err *model.AppError

	if _, err = p.API.CreatePost(post); err != nil {
		p.API.LogInfo("Unable to post to channel")
		return
	}

}
