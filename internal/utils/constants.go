package utils

const VERSION = "v1"
const REGISTRATION_PATH = "/envdash/" + VERSION + "/registrations/"
const AUTHENTICATION_PATH = "/envdash/" + VERSION + "/auth"
const MAXAPIKEYS = 5
const NOTIFICATION_PATH = "/envdash/" + VERSION + "/notifications"

//notifications
var VALIDEVENTS []string = []string{"REGISTER", "CHANGE", "DELETE", "INVOKE", "THRESHOLD"}
