package main

import (
	"fmt"
	"github.com/AgoraIO-Community/go-tokenbuilder/rtctokenbuilder"
	"github.com/AgoraIO-Community/go-tokenbuilder/rtmtokenbuilder"
	"github.com/spf13/viper"
	"log"
	"strconv"

	"github.com/julienschmidt/httprouter"

	"net/http"
	"time"
)



func main() {


	router := httprouter.New()

	router.GET("/",Index)
	router.GET("/rtc/:channelName/:role/:tokentype/:uid/:expiry/", getRtcToken)
	router.GET("/rtm/:uid/", getRtmToken)
	router.GET("/rte/:channelName/:role/:tokentype/:uid", getBothTokens)


	err := http.ListenAndServe(":"+"7000", router) //Launch the app, visit localhost:8000/api
	if err != nil {
		fmt.Print(err)
	}

}


func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")

	fmt.Println(time.Now().Format(time.RFC850))


	//u.SendEmail("murainoy@yahoo.com",u.EmailTemplate("223445", "test"))


}


func getRtcToken(w http.ResponseWriter, c *http.Request, r httprouter.Params) {
	log.Printf("rtc token\n")
	// get param values
	channelName, tokentype, uidStr, role, expireTimestamp, err := parseRtcParams(r)

	if err != nil {
		resp := Message(false, err.Error())
		Responds(w, resp)
		return
	}


	rtcToken, tokenErr := generateRtcToken(channelName, uidStr, tokentype, role, expireTimestamp)

	if tokenErr != nil {
		resp := Message(false, tokenErr.Error())
		Responds(w, resp)
		return
	} else {
		log.Println("RTC Token generated")
		resp := Message(false, "RTC Token generated")
		resp["data"] = rtcToken
		Responds(w, resp)
		return
	}
}

func getRtmToken(w http.ResponseWriter, r *http.Request, c httprouter.Params) {

	InitializeViper()

	appID := viper.GetString("AppID")
	appCertificate := viper.GetString("AppCertificate")

	log.Printf("rtm token\n")
	// get param values
	uidStr, expireTimestamp, err := parseRtmParams(c)

	if err != nil {
		resp := Message(false, err.Error())
		Responds(w, resp)
		return
	}

	rtmToken, tokenErr := rtmtokenbuilder.BuildToken(appID, appCertificate, uidStr, rtmtokenbuilder.RoleRtmUser, expireTimestamp)

	if tokenErr != nil {
		resp := Message(false, tokenErr.Error())
		Responds(w, resp)
		return
	} else {
		log.Println("RTM Token generated")

		resp := Message(false,"RTM Token generated")
		resp["rtmToken"]= rtmToken

		Responds(w, resp)
		return
	}
}

func getBothTokens(w http.ResponseWriter, r *http.Request, c httprouter.Params) {

	InitializeViper()

	appID := viper.GetString("AppID")
	appCertificate := viper.GetString("AppCertificate")

	log.Printf("dual token\n")
	// get rtc param values
	channelName, tokentype, uidStr, role, expireTimestamp, rtcParamErr := parseRtcParams(c)

	if rtcParamErr != nil {

		resp := Message(false,"Error Generating RTC token: " + rtcParamErr.Error())
		Responds(w, resp)
		return

	}
	// generate the rtcToken
	rtcToken, rtcTokenErr := generateRtcToken(channelName, uidStr, tokentype, role, expireTimestamp)
	// generate rtmToken
	rtmToken, rtmTokenErr := rtmtokenbuilder.BuildToken(appID, appCertificate, uidStr, rtmtokenbuilder.RoleRtmUser, expireTimestamp)

	if rtcTokenErr != nil {
		log.Println(rtcTokenErr) // token failed to generate
		resp := Message(false,"Error Generating RTC token - " + rtcTokenErr.Error())
		Responds(w, resp)
		return

	} else if rtmTokenErr != nil {

		log.Println(rtcTokenErr) // token failed to generate
		resp := Message(false,"Error Generating RTC token -" + rtmTokenErr.Error())
		Responds(w, resp)
		return

	} else {

		log.Println(rtcTokenErr) // token failed to generate
		resp := Message(false,"RTC Token generated")
		resp["rtcToken"]= rtcToken
		resp["rtmToken"]= rtmToken
		Responds(w, resp)
		return

	}
}

func parseRtcParams(c httprouter.Params) (channelName, tokentype, uidStr string, role rtctokenbuilder.Role, expireTimestamp uint32, err error) {

	channelName = c.ByName("channelName")
	roleStr := c.ByName("role")
	tokentype = c.ByName("tokentype")
	uidStr = c.ByName("uid")
	expireTime := c.ByName("expiry")

	if roleStr == "publisher" {
		role = rtctokenbuilder.RolePublisher
	} else {
		role = rtctokenbuilder.RoleSubscriber
	}

	expireTime64, parseErr := strconv.ParseUint(expireTime, 10, 64)
	if parseErr != nil {
		// if string conversion fails return an error
		err = fmt.Errorf("failed to parse expireTime: %s, causing error: %s", expireTime, parseErr)
	}

	// set timestamps
	expireTimeInSeconds := uint32(expireTime64)
	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp = currentTimestamp + expireTimeInSeconds

	return channelName, tokentype, uidStr, role, expireTimestamp, err
}

func parseRtmParams(c httprouter.Params) (uidStr string, expireTimestamp uint32, err error) {

	// get param values
	uidStr = c.ByName("uid")
	expireTime := c.ByName("expiry")

	expireTime64, parseErr := strconv.ParseUint(expireTime, 10, 64)
	if parseErr != nil {
		// if string conversion fails return an error
		err = fmt.Errorf("failed to parse expireTime: %s, causing error: %s", expireTime, parseErr)
	}

	// set timestamps
	expireTimeInSeconds := uint32(expireTime64)
	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp = currentTimestamp + expireTimeInSeconds

	// check if string conversion fails
	return uidStr, expireTimestamp, err
}

func generateRtcToken(channelName, uidStr, tokentype string, role rtctokenbuilder.Role, expireTimestamp uint32) (rtcToken string, err error) {

	InitializeViper()

	appID := viper.GetString("AppID")
	appCertificate := viper.GetString("AppCertificate")

	if tokentype == "userAccount" {
		log.Printf("Building Token with userAccount: %s\n", uidStr)
		rtcToken, err = rtctokenbuilder.BuildTokenWithUserAccount(appID, appCertificate, channelName, uidStr, role, expireTimestamp)
		return rtcToken, err

	} else if tokentype == "uid" {
		uid64, parseErr := strconv.ParseUint(uidStr, 10, 64)
		// check if conversion fails
		if parseErr != nil {
			err = fmt.Errorf("failed to parse uidStr: %s, to uint causing error: %s", uidStr, parseErr)
			return "", err
		}

		uid := uint32(uid64) // convert uid from uint64 to uint 32
		log.Printf("Building Token with uid: %d\n", uid)
		rtcToken, err = rtctokenbuilder.BuildTokenWithUID(appID, appCertificate, channelName, uid, role, expireTimestamp)
		return rtcToken, err

	} else {
		err = fmt.Errorf("failed to generate RTC token for Unknown Tokentype: %s", tokentype)
		log.Println(err)
		return "", err
	}
}