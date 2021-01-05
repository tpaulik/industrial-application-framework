# alarmlogger package

**import "github.com/nokia/industrial-application-framework/alarmlogger"**

**alarmlogger** is a simple _go_ module for producing NDAC Application alarm log in proper format.

## Usage
1. Import this module.
2. Create an instance of alarmlogger.AlarmDetails which has the following fields currently:
	```
	type AlarmDetails struct {
		Name       string `json:"name"`
		ID         string `json:"id"`
		Severity   AlarmSeverity `json:"severity"`
		Text       string `json:"text"`
		State      int    `json:"state"`
		Visibility AlarmVisibility `json:"visibility,omitempty"`
		SubDN      string `json:"subdn,omitempty"`
	}

	| Field      | Description                                                                                         |
	|------------|-----------------------------------------------------------------------------------------------------|
	| Name       | Alarm Name                                                                                          |
	|            | Recommendation: Short name in UpperCamelCaseFormat                                                  |
	| ID         | Identification for the alarm aside from its name (recommended to be convertible to int)             |
	|            | (Recommended to be convertible to int, as common practice)                                          |
	| Severity   | Severity of the problem                                                                             |
	|            | Recommended values: see Constants section                                                           |
	| Text       | Short description of the problem                                                                    |
	| State      | Status of the problem whether active or already cleared<br>Recommended values: 0/1 (Cleared/Active) |
	| Visibility | Visibility of the alarm                                                                             |
	|            | Recommended values: see Constants section                                                           |
	| SubDN      | Distinguished Name, if any, that is internal to application or service that is reporting the alarm  |
	|            | It can pertain to a module, package, etc of an application or service                               |
	|            | If in use, the same SubDN should be used for raise and clear alarm                                  |
	|            | To follow CAM's DN format, apps using this should provide value with the following format:          |
	|            |     "/<SOMEKEY>-<Distinguish Name>"                                                                 |
	|            |     ex. "/MODULE-servicesubmodule"                                                                  |

3. This module has the following function definitions for raising and clearing alarm, respectively:
	```
	func RaiseAlarm(logtype string, alarm *AlarmDetails)
	func ClearAlarm(logtype string, alarm *AlarmDetails)
	```
	logtype parameter is a label for alarm filtering. See Constants section for recommended values.

    Raise an alarm:
    ```
    alarmlogger.RaiseAlarm(alarmlogger.AppAlarm, &alarmlogger.AlarmDetails{Name: "AlarmName", ID: "123", Severity: alarmlogger.Minor, Text: "any", SubDN: "/MODULE-servicesubmodule"})
    ```
	Output:
	```
	{"ts":"2020-06-22T06:47:22.637+0200","log_type":"APP_ALARM","alarm":{"name":"AlarmName","id":"123","severity":"MINOR","text":"any","state":1,"visibility":"GLOBAL","subdn":"/MODULE-servicesubmodule"}}
	```
    RaiseAlarm sets the alarmlogger.AlarmDetails.State to 1 and alarmlogger.AlarmDetails.Visibility to "GLOBAL" if not set.

    Clear an alarm:
    ```
    alarmlogger.ClearAlarm(alarmlogger.AppAlarm, &alarmlogger.AlarmDetails{Name: "AlarmName", ID: "123", Severity: alarmlogger.Minor, Text: "any", SubDN: "/MODULE-servicesubmodule"})
    ```
	Output:
	```
	{"ts":"2020-06-22T06:47:22.637+0200","log_type":"APP_ALARM","alarm":{"name":"AlarmName","id":"123","severity":"MINOR","text":"any","state":0,"subdn":"/MODULE-servicesubmodule"}}
	```
    ClearAlarm sets the alarmlogger.AlarmDetails.State to 0.

## Constants
```
const (
	// log type
	AppAlarm   LogType = "APP_ALARM"   // for NDAC Applications
	AppFwAlarm LogType = "APPFW_ALARM" // for NDAC APPFW components

	// severity
	Minor    AlarmSeverity = "MINOR"
	Major    AlarmSeverity = "MAJOR"
	Critical AlarmSeverity = "CRITICAL"
	Warning  AlarmSeverity = "WARNING"
	Info     AlarmSeverity = "INFO"

	// visibility
	Global 	   AlarmVisibility = "GLOBAL"		// default, visible "anywhere"
	Operations AlarmVisibility = "OPERATIONS"   // not visible in C-UI (of NDAC)
)
```
