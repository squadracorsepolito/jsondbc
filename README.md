# Canconv

Simple go cli to convert CAN models defined in JSON to dbc.

## Installation

Download the binary in the release section

## Usage

```
canconv convert --in my_model.json --out my_dbc_model.dbc
```

## CAN Model

| field              | type                 | description                                                                     |
| ------------------ | -------------------- | ------------------------------------------------------------------------------- |
| version            | string               | The version of the CAN model                                                    |
| bus_speed          | number               | The bus spedd of the CAN model                                                  |
| nodes              | map[string]Node      | A map containing the nodes as value and the node names as key                   |
| general_attributes | map[string]Attribute | A map containing the general attributes as value and the attribute names as key |
| node_attributes    | map[string]Attribute | A map containing the node attributes as value and the attribute names as key    |
| message_attributes | map[string]Attribute | A map containing the message attributes as value and the attribute names as key |
| signal_attributes  | map[string]Attribute | A map containing the signal attributes as value and the attribute names as key  |
| messages           | map[string]Message   | A map containig the messages as value and the message names as key              |

### Node

| field       | type   | description            |
| ----------- | ------ | ---------------------- |
| description | string | The node's description |

### Message

| field       | type              | description                                                         | required |
| ----------- | ----------------- | ------------------------------------------------------------------- | -------- |
| id          | number            | The message's id in decimal                                         | true     |
| description | string            | The message's description                                           | false    |
| length      | number            | The message's length (bytes count)                                  | true     |
| sender      | string            | The message's sender name                                           | false    |
| signals     | map[string]Signal | A map containing the message's signals, with the signal name as key | true     |

### Signal

| field       | type              | description                                                                                                     | required                  | default |
| ----------- | ----------------- | --------------------------------------------------------------------------------------------------------------- | ------------------------- | ------- |
| start_bit   | number            | The signal's start bit                                                                                          | true                      |
| size        | number            | The signal's size (bits count)                                                                                  | true                      |
| description | string            | The signal's description                                                                                        | false                     |
| big_endian  | boolean           | The signal's byte order                                                                                         | false                     | false   |
| signed      | boolean           | The signal's value type                                                                                         | false                     | false   |
| receivers   | string[]          | The signal's receivers list                                                                                     | false                     |
| scale       | number            | The signal's scale                                                                                              | false                     | 1       |
| offset      | number            | The signal's offset                                                                                             | false                     | 0       |
| min         | number            | The signal's minimum value                                                                                      | false                     | 0       |
| max         | number            | The signal's maximum value                                                                                      | true                      |
| bitmap      | map[string]number | A map with key the _uman readable_ name corrisponding to a signal's vlue                                        | false                     |
| mux_group   | map[string]Signal | A map with key the name of a multiplexed signal and a Signal as value. If set, the signal becomes a multiplexor | false                     |
| mux_switch  | number            | The value a multiplexor signal as to be in order to map to the multiplexed signal                               | Only if part of mux_group |
