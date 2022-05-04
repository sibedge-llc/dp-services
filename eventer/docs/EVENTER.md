# Eventer

Eventer is the generator tool to produce series of events for the kafka/postgres.

The main idea is to generate events of any topology and not bother to any scheme. 

Producing events is different for kafka and postgres. `postgres` implies to have the clear scheme for upsert operation whereas `kafka` doesn't. 

## Event
Event is the randonly generated data in json form. generate random data used [jsonnet](https://jsonnet.org/) with extension of custom methods to bring random bits

Here is the custom methods that can be used into jsonnet to describe vary of random values

```jsonnet
// (string) dataset name supplied from the request
local get_dataset = std.native('get_dataset');

// (string) instance id supplied by config of the service
local get_instance_id = std.native('get_instance_id');

// get one of the space separated string values
local get_one_of = std.native('get_one_of');

// random timestamp within given time frame defined by start/end values with step 
local get_timestamp = std.native('get_timestamp');

// random integer value in range
local get_integer = std.native('get_integer');

// random number (floating point) value in range
local get_number = std.native('get_number');

// returns json with full of random data (see below for details)
local get_rand_data = std.native('get_rand_data');

// get random user agent value
local get_rand_user_agent = std.native('get_rand_user_agent');
```

### get_rand_data()
```jsonnet
get_rand_data()["first_name"]
```

| Key                  |
| -------------------- |
| lat                  |
| long                 |
| cc_number            |
| cc_type              |
| email                |
| domain_name          |
| ipv4                 |
| ipv6                 |
| password             |
| jwt                  |
| phone_number         |
| mac_address          |
| url                  |
| username             |
| toll_free_number     |
| e_164_phone_number   |
| first_name           |
| last_name            |
| name                 |
| unix_time            |
| date                 |
| time                 |
| month_name           |
| year                 |
| day_of_month         |
| timestamp            |
| century              |
| timezone             |
| time_period          |
| word                 |
| sentence             |
| paragraph            |
| currency             |
| amount               |
| amount_with_currency |
| uuid_hyphenated      |
| uuid_digit           |
| opayment_method      |
| id                   |
| price                |
| number               |


### Event.jsonnet

```jsonnet
local get_dataset = std.native('get_dataset');
local get_instance_id = std.native('get_instance_id');
local get_one_of = std.native('get_one_of');
local get_timestamp = std.native('get_timestamp');
local get_integer = std.native('get_integer');
local get_number = std.native('get_number');
local get_rand_data = std.native('get_rand_data');
local get_rand_user_agent = std.native('get_rand_user_agent');

{
    id: get_integer(1, 1000),
    time: get_timestamp("2021-12-01", "2021-12-31", "1h"),
    group_id: get_instance_id("dummy_instance_id"),
    dataset: get_dataset("dummy_dataset"),
    country: get_one_of("ru, us, jp, cn"),
    name: get_rand_data()["first_name"],
    session_id: get_integer(1, 300),
    page_id: get_integer(1, 15),
    user_agent: get_rand_user_agent(),
    event: {
        type: get_one_of("click, open_page, button_click"),
        data: {
            [if self.type == "click" then self.type else null]: {
                x: get_number(1, 1000),
                y: get_number(1, 1000),
            },
            [if self.type == "button_click" then self.type else null]: {
                button_id: get_integer(1, 100),
            },
            [if self.type == "open_page" then self.type else null]: {
                page_id: $["page_id"],
            },
           
        },
    },
}
```

## Kafka
`kafka` event producing is simple as ensure topic is exists. Generator tries to create topic with requested name and start
sending events in json format

## Postgres
`postgres` event producing is a bit tricky. Once the new request to start generate events, generator tries to ensure table with given scheme exist and matched to the event data. This causes the missed table is created and missed columns are added accordingly. So no any specific setup is requested in db.


## Usage

### Add new generator (kafka)

Add 2 generators to produce 10 events with 10 and 8 seconds intervals that send messages to topics `moo` and `boo`

```shell
curl --location --request POST 'localhost:9099/generator/add' \
--header 'Content-Type: application/json' \
--data-raw '{
    "destinations": [
        {
            "id": "d1",
            "type": "kafka",
            "kafka": {
                "topic": "boo"
            }
        },
        {
            "id": "d2",
            "type": "kafka",
            "kafka": {
                "topic": "moo"
            }
        }
    ],
    "events": [
        {
            "id": "e1",
            "dataset": "crazy_airflow",
            "schema": "bG9jYWwgZ2V0X2RhdGFzZXQgPSBzdGQubmF0aXZlKCdnZXRfZGF0YXNldCcpOwpsb2NhbCBnZXRfaW5zdGFuY2VfaWQgPSBzdGQubmF0aXZlKCdnZXRfaW5zdGFuY2VfaWQnKTsKbG9jYWwgZ2V0X29uZV9vZiA9IHN0ZC5uYXRpdmUoJ2dldF9vbmVfb2YnKTsKbG9jYWwgZ2V0X3RpbWVzdGFtcCA9IHN0ZC5uYXRpdmUoJ2dldF90aW1lc3RhbXAnKTsKbG9jYWwgZ2V0X2ludGVnZXIgPSBzdGQubmF0aXZlKCdnZXRfaW50ZWdlcicpOwpsb2NhbCBnZXRfbnVtYmVyID0gc3RkLm5hdGl2ZSgnZ2V0X251bWJlcicpOwpsb2NhbCBnZXRfcmFuZF9kYXRhID0gc3RkLm5hdGl2ZSgnZ2V0X3JhbmRfZGF0YScpOwpsb2NhbCBnZXRfcmFuZF91c2VyX2FnZW50ID0gc3RkLm5hdGl2ZSgnZ2V0X3JhbmRfdXNlcl9hZ2VudCcpOwoKewogICAgaWQ6IGdldF9pbnRlZ2VyKDEsIDEwMDApLAogICAgdGltZTogZ2V0X3RpbWVzdGFtcCgiMjAyMS0xMi0wMSIsICIyMDIxLTEyLTMxIiwgIjFoIiksCiAgICBncm91cF9pZDogZ2V0X2luc3RhbmNlX2lkKCJkdW1teV9pbnN0YW5jZV9pZCIpLAogICAgZGF0YXNldDogZ2V0X2RhdGFzZXQoImR1bW15X2RhdGFzZXQiKSwKICAgIGNvdW50cnk6IGdldF9vbmVfb2YoInJ1LCB1cywganAsIGNuIiksCiAgICBuYW1lOiBnZXRfcmFuZF9kYXRhKClbImZpcnN0X25hbWUiXSwKICAgIHNlc3Npb25faWQ6IGdldF9pbnRlZ2VyKDEsIDMwMCksCiAgICBwYWdlX2lkOiBnZXRfaW50ZWdlcigxLCAxNSksCiAgICB1c2VyX2FnZW50OiBnZXRfcmFuZF91c2VyX2FnZW50KCksCiAgICBldmVudDogewogICAgICAgIHR5cGU6IGdldF9vbmVfb2YoImNsaWNrLCBvcGVuX3BhZ2UsIGJ1dHRvbl9jbGljayIpLAogICAgICAgIGRhdGE6IHsKICAgICAgICAgICAgW2lmIHNlbGYudHlwZSA9PSAiY2xpY2siIHRoZW4gc2VsZi50eXBlIGVsc2UgbnVsbF06IHsKICAgICAgICAgICAgICAgIHg6IGdldF9udW1iZXIoMSwgMTAwMCksCiAgICAgICAgICAgICAgICB5OiBnZXRfbnVtYmVyKDEsIDEwMDApLAogICAgICAgICAgICB9LAogICAgICAgICAgICBbaWYgc2VsZi50eXBlID09ICJidXR0b25fY2xpY2siIHRoZW4gc2VsZi50eXBlIGVsc2UgbnVsbF06IHsKICAgICAgICAgICAgICAgIGJ1dHRvbl9pZDogZ2V0X2ludGVnZXIoMSwgMTAwKSwKICAgICAgICAgICAgfSwKICAgICAgICAgICAgW2lmIHNlbGYudHlwZSA9PSAib3Blbl9wYWdlIiB0aGVuIHNlbGYudHlwZSBlbHNlIG51bGxdOiB7CiAgICAgICAgICAgICAgICBwYWdlX2lkOiAkWyJwYWdlX2lkIl0sCiAgICAgICAgICAgIH0sCiAgICAgICAgICAgCiAgICAgICAgfSwKICAgIH0sCn0=",
            "count": 10,
            "interval": "10s"
        },
        {
            "id": "e2",
            "dataset": "crazy_airflow",
            "schema": "bG9jYWwgZ2V0X2RhdGFzZXQgPSBzdGQubmF0aXZlKCdnZXRfZGF0YXNldCcpOwpsb2NhbCBnZXRfaW5zdGFuY2VfaWQgPSBzdGQubmF0aXZlKCdnZXRfaW5zdGFuY2VfaWQnKTsKbG9jYWwgZ2V0X29uZV9vZiA9IHN0ZC5uYXRpdmUoJ2dldF9vbmVfb2YnKTsKbG9jYWwgZ2V0X3RpbWVzdGFtcCA9IHN0ZC5uYXRpdmUoJ2dldF90aW1lc3RhbXAnKTsKbG9jYWwgZ2V0X2ludGVnZXIgPSBzdGQubmF0aXZlKCdnZXRfaW50ZWdlcicpOwpsb2NhbCBnZXRfbnVtYmVyID0gc3RkLm5hdGl2ZSgnZ2V0X251bWJlcicpOwpsb2NhbCBnZXRfcmFuZF9kYXRhID0gc3RkLm5hdGl2ZSgnZ2V0X3JhbmRfZGF0YScpOwpsb2NhbCBnZXRfcmFuZF91c2VyX2FnZW50ID0gc3RkLm5hdGl2ZSgnZ2V0X3JhbmRfdXNlcl9hZ2VudCcpOwoKewogICAgaWQ6IGdldF9pbnRlZ2VyKDEsIDEwMDApLAogICAgaXA6IGdldF9yYW5kX2RhdGEoKVsiaXB2NCJdLAogICAgdXNlcl9hZ2VudDogZ2V0X3JhbmRfdXNlcl9hZ2VudCgpLAogICAgcHJhc2U6IGdldF9vbmVfb2YoImhpLCBtb28sIGRvb2gsIG9vcHMsIHdoZW5ldmVyIiksCn0=",
            "count": 10,
            "interval": "8s"
        }
    ],
    "schedules": [
        {
            "destination_id": "d1",
            "event_id": "e1"
        },
        {
            "destination_id": "d2",
            "event_id": "e2"
        }
    ] 
}'

```

`schema` properties is base64 encoded jsonnet data that describes event data. Thus value can be generated with `base64` command line tool. 

```shell
echo "... jsonnet contents... " | base64 
```

or 
```shell
cat file.jsonnet | base64 
```

### Add new generator (postgress)

Add 2 generators that upsert the rows with randomly generated data to tables `moo` and `boo` with time interval 10 and 8 seconds

```shell
curl --location --request POST 'localhost:9099/generator/add' \
--header 'Content-Type: application/json' \
--data-raw '{
    "destinations": [
        {
            "id": "d1",
            "type": "postgres",
            "postgres": {
                "table": "boo"
            }
        },
        {
            "id": "d2",
            "type": "postgres",
            "postgres": {
                "table": "moo"
            }
        }
    ],
    "events": [
        {
            "id": "e1",
            "dataset": "crazy_airflow",
            "schema": "bG9jYWwgZ2V0X2RhdGFzZXQgPSBzdGQubmF0aXZlKCdnZXRfZGF0YXNldCcpOwpsb2NhbCBnZXRfaW5zdGFuY2VfaWQgPSBzdGQubmF0aXZlKCdnZXRfaW5zdGFuY2VfaWQnKTsKbG9jYWwgZ2V0X29uZV9vZiA9IHN0ZC5uYXRpdmUoJ2dldF9vbmVfb2YnKTsKbG9jYWwgZ2V0X3RpbWVzdGFtcCA9IHN0ZC5uYXRpdmUoJ2dldF90aW1lc3RhbXAnKTsKbG9jYWwgZ2V0X2ludGVnZXIgPSBzdGQubmF0aXZlKCdnZXRfaW50ZWdlcicpOwpsb2NhbCBnZXRfbnVtYmVyID0gc3RkLm5hdGl2ZSgnZ2V0X251bWJlcicpOwpsb2NhbCBnZXRfcmFuZF9kYXRhID0gc3RkLm5hdGl2ZSgnZ2V0X3JhbmRfZGF0YScpOwpsb2NhbCBnZXRfcmFuZF91c2VyX2FnZW50ID0gc3RkLm5hdGl2ZSgnZ2V0X3JhbmRfdXNlcl9hZ2VudCcpOwoKewogICAgaWQ6IGdldF9pbnRlZ2VyKDEsIDEwMDApLAogICAgdGltZTogZ2V0X3RpbWVzdGFtcCgiMjAyMS0xMi0wMSIsICIyMDIxLTEyLTMxIiwgIjFoIiksCiAgICBncm91cF9pZDogZ2V0X2luc3RhbmNlX2lkKCJkdW1teV9pbnN0YW5jZV9pZCIpLAogICAgZGF0YXNldDogZ2V0X2RhdGFzZXQoImR1bW15X2RhdGFzZXQiKSwKICAgIGNvdW50cnk6IGdldF9vbmVfb2YoInJ1LCB1cywganAsIGNuIiksCiAgICBuYW1lOiBnZXRfcmFuZF9kYXRhKClbImZpcnN0X25hbWUiXSwKICAgIHNlc3Npb25faWQ6IGdldF9pbnRlZ2VyKDEsIDMwMCksCiAgICBwYWdlX2lkOiBnZXRfaW50ZWdlcigxLCAxNSksCiAgICB1c2VyX2FnZW50OiBnZXRfcmFuZF91c2VyX2FnZW50KCksCiAgICBldmVudDogewogICAgICAgIHR5cGU6IGdldF9vbmVfb2YoImNsaWNrLCBvcGVuX3BhZ2UsIGJ1dHRvbl9jbGljayIpLAogICAgICAgIGRhdGE6IHsKICAgICAgICAgICAgW2lmIHNlbGYudHlwZSA9PSAiY2xpY2siIHRoZW4gc2VsZi50eXBlIGVsc2UgbnVsbF06IHsKICAgICAgICAgICAgICAgIHg6IGdldF9udW1iZXIoMSwgMTAwMCksCiAgICAgICAgICAgICAgICB5OiBnZXRfbnVtYmVyKDEsIDEwMDApLAogICAgICAgICAgICB9LAogICAgICAgICAgICBbaWYgc2VsZi50eXBlID09ICJidXR0b25fY2xpY2siIHRoZW4gc2VsZi50eXBlIGVsc2UgbnVsbF06IHsKICAgICAgICAgICAgICAgIGJ1dHRvbl9pZDogZ2V0X2ludGVnZXIoMSwgMTAwKSwKICAgICAgICAgICAgfSwKICAgICAgICAgICAgW2lmIHNlbGYudHlwZSA9PSAib3Blbl9wYWdlIiB0aGVuIHNlbGYudHlwZSBlbHNlIG51bGxdOiB7CiAgICAgICAgICAgICAgICBwYWdlX2lkOiAkWyJwYWdlX2lkIl0sCiAgICAgICAgICAgIH0sCiAgICAgICAgICAgCiAgICAgICAgfSwKICAgIH0sCn0=",
            "count": 10,
            "interval": "10s"
        },
        {
            "id": "e2",
            "dataset": "crazy_airflow",
            "schema": "bG9jYWwgZ2V0X2RhdGFzZXQgPSBzdGQubmF0aXZlKCdnZXRfZGF0YXNldCcpOwpsb2NhbCBnZXRfaW5zdGFuY2VfaWQgPSBzdGQubmF0aXZlKCdnZXRfaW5zdGFuY2VfaWQnKTsKbG9jYWwgZ2V0X29uZV9vZiA9IHN0ZC5uYXRpdmUoJ2dldF9vbmVfb2YnKTsKbG9jYWwgZ2V0X3RpbWVzdGFtcCA9IHN0ZC5uYXRpdmUoJ2dldF90aW1lc3RhbXAnKTsKbG9jYWwgZ2V0X2ludGVnZXIgPSBzdGQubmF0aXZlKCdnZXRfaW50ZWdlcicpOwpsb2NhbCBnZXRfbnVtYmVyID0gc3RkLm5hdGl2ZSgnZ2V0X251bWJlcicpOwpsb2NhbCBnZXRfcmFuZF9kYXRhID0gc3RkLm5hdGl2ZSgnZ2V0X3JhbmRfZGF0YScpOwpsb2NhbCBnZXRfcmFuZF91c2VyX2FnZW50ID0gc3RkLm5hdGl2ZSgnZ2V0X3JhbmRfdXNlcl9hZ2VudCcpOwoKewogICAgaWQ6IGdldF9pbnRlZ2VyKDEsIDEwMDApLAogICAgaXA6IGdldF9yYW5kX2RhdGEoKVsiaXB2NCJdLAogICAgdXNlcl9hZ2VudDogZ2V0X3JhbmRfdXNlcl9hZ2VudCgpLAogICAgcHJhc2U6IGdldF9vbmVfb2YoImhpLCBtb28sIGRvb2gsIG9vcHMsIHdoZW5ldmVyIiksCn0=",
            "count": 10,
            "interval": "8s"
        }
    ],
    "schedules": [
        {
            "destination_id": "d1",
            "event_id": "e1"
        },
        {
            "destination_id": "d2",
            "event_id": "e2"
        }
    ] 
}'
```

Generator supports followed time interval notations 
`"ns", "us" (or "µs"), "ms", "s", "m", "h"`

Once generators are added the generator is is returned. 
Generator is smart to recognise similar schemas (for instance you may change keys order in jsonnet) or use the same scheme and call add as many times as you want. Only one generator instance is added to prevent generators hell that produce similar events chaotically.

### Get generator status

```shell
curl --location --request GET 'localhost:9099/generator/status?id=12594183362362990045' \
--header 'Content-Type: application/json'
```

### Stop and delete generator

```shell
curl --location --request POST 'localhost:9099/generator/remove?id=11168030917768196602' \
--header 'Content-Type: application/json' \
--data-raw '{
    "id": "17828440514488382438"
}'
```