# Notification

###  Advanced Task
API keys determine who can register webhooks and what level of access they have when deleting notifications or listing all stored notifications.


## Parameters:

| Parameter      | Type  | Description                       |
|:-----------|:------------|:----------------------------    |
| `url`     | `string`    | *Required* URL to POST to when the event fires                      |
| `country`    | `string`    | *Optional*  2 letter country code ([lookup here](https://datahub.io/core/country-list)). Omit or leave empty to match all countries|
| `event`     | `string`    | *Required* One of: `REGISTER`, `CHANGE`, `DELETE`, `INVOKE`, `THRESHOLD`                    |


| Event | Triggered when… |
|-------|-----------------|
| `REGISTER` | A new dashboard configuration is registered (`POST /registrations/`) |
| `CHANGE` | A configuration is updated (`PUT` or `PATCH /registrations/{id}`) |
| `DELETE` | A configuration is deleted (`DELETE /registrations/{id}`) |
| `INVOKE` | A populated dashboard is retrieved (`GET /dashboards/{id}`) |
| `THRESHOLD` | A live measured value crosses a user-defined threshold during dashboard retrieval |

For threshold notificatoins:
| Parameter      | Type  | Description                       |
|:-----------|:------------|:----------------------------    |
| `threshold.field`  | `string`    | *Required* Field to monitor: `pm25` \| `pm10` \| `temperature` \| `precipitation`                   |
| `threshold.operator`| `Comparison operator`    | *Required*  Comparison operator: `>`, `<`, `>=`, `<=`, `==` |
| `threshold.value`  | `string`    | *Required* Numeric threshold value        |


### Design desisions
We decided to fully implement authenification in notifications, your notifications are tightly tied to your account

### How are we sending webhooks?

#### Lifecycle
Each time a registration or dashboard changes or deletes etc we check all notificaitons if they match a lifecycle event.

#### Threshold
Each time a dashboard is retrieved (GET /dashboards/{id}), the live measured values are compared against all registered threshold webhooks. If a condition is met, a POST is sent to the registered URL.