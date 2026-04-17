package utils

const VERSION = "v1"
const REGISTRATION_PATH = "/envdash/" + VERSION + "/registrations/"
const DASHBOARD_PATH = "/envdash/" + VERSION + "/dashboards/"
const STATUS_PATH = "/envdash/" + VERSION + "/status"

const AUTHENTICATION_PATH = "/envdash/" + VERSION + "/auth"
const NOTIFICATION_PATH = "/envdash/" + VERSION + "/notifications"

// Authentication
const MAXAPIKEYS = 5
const MAXATTEMPTSFORKEYGENERATION = 10
const STARTOFUSERAPI = "sk-envdash-"

// notifications
var VALIDEVENTS []string = []string{"REGISTER", "CHANGE", "DELETE", "INVOKE", "THRESHOLD"}
var VALIDTHRESHOLDS []string = []string{"PM25", "PM10", "TEMPERATURE", "PRECIPITATION"}
var VALIDOPERATORS []string = []string{">", "<", ">=", "<=", "=="}

const LONGEST_COUNTRYNAME = 56
const SHORTEST_COUNTRYNAME = 4
const ISOCODE_LENGTH = 2

const CURRENCYCODE_LENGTH = 3
const TARGETCURRENCIES_MAX_LENGTH = 10

const CURRENCY_URL = "http://129.241.150.113:9090/currency/AED"
const COUNTRY_AND_ISO_URL = "http://129.241.150.113:8080/v3.1/all?fields=name,cca2"
