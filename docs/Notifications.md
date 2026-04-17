# Notification

###  Advanced Task
Added authentication with Firestore as the storage backend. API keys determine who can register webhooks and what level of access they have when deleting notifications or listing all stored notifications.


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
We decided to allow everyone to look up notification IDs, regardless of whether they own them or not. This is intended to help combat spam. When you receive a webhook, the notification ID is included as well. It would be unwise to prevent the recipient from reading the notification they received.

### How are we sending webhooks?

#### Lifecycle
Each time a registration or dashboard changes or deletes etc we check all notificaitons if they match a lifecycle event.

#### Threshold
We send a POST request for each time the cache is updated. This is the time we check each webhook for correlation.