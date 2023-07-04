# jsondbc

Simple go cli to convert CAN models in JSON and dbc.

## Installation

Download the binary in the release section

## Usage

Converting from json to dbc:

```
jsondbc convert --in my_model.json --out my_dbc_model.dbc
```

Converting from dbc to json:

```
jsondbc convert --in my_model.dbc --out my_dbc_model.json
```

## CAN Model

| field              | type                 | description                                                                     |
| ------------------ | -------------------- | ------------------------------------------------------------------------------- |
| version            | string               | The version of the CAN model                                                    |
| baudrate           | number               | The baud rate of the CAN model                                                  |
| nodes              | map[string]Node      | A map containing the nodes as value and the node names as key                   |
| general_attributes | map[string]Attribute | A map containing the general attributes as value and the attribute names as key |
| node_attributes    | map[string]Attribute | A map containing the node attributes as value and the attribute names as key    |
| message_attributes | map[string]Attribute | A map containing the message attributes as value and the attribute names as key |
| signal_attributes  | map[string]Attribute | A map containing the signal attributes as value and the attribute names as key  |
| messages           | map[string]Message   | A map containig the messages as value and the message names as key              |

### Attribute

| field  | type            | description                   |
| ------ | --------------- | ----------------------------- |
| int    | AttributeInt    | Set's the attribute as int    |
| string | AttributeString | Set's the attribute as string |
| float  | AttributeFloat  | Set's the attribute as float  |
| enum   | AttributeEnum   | Set's the attribute as enum   |

### AttributeInt

| field   | type   | description                        |
| ------- | ------ | ---------------------------------- |
| default | number | The attribute's default value      |
| from    | number | The attribute's lower bound value  |
| to      | number | The attribute's uppuer bound value |

### AttributeString

| field   | type   | description                   |
| ------- | ------ | ----------------------------- |
| default | string | The attribute's default value |

### AttributeFloat

| field   | type   | description                        |
| ------- | ------ | ---------------------------------- |
| default | number | The attribute's default value      |
| from    | number | The attribute's lower bound value  |
| to      | number | The attribute's uppuer bound value |

### AttributeEnum

| field   | type     | description                             |
| ------- | -------- | --------------------------------------- |
| default | string   | The attribute's default value           |
| values  | string[] | The list of possible attribute's values |

### Node

| field       | type           | description                                                                                                                                                                                    |
| ----------- | -------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| description | string         | The node's description                                                                                                                                                                         |
| attributes  | map[string]any | A map with key the attribute name and a value to assign as map's value. The value must be an int if the attribute is of type int, string for type string, an enum value (string) for type enum |

### Message

| field       | type              | description                                                                                                                                                                                    | required |
| ----------- | ----------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| id          | number            | The message's id in decimal                                                                                                                                                                    | true     |
| period_ms   | number            | The message's period in ms. If set, it creates an int attribute named "MsgPeriodMS" with the corrisponding period                                                                              | false    |
| description | string            | The message's description                                                                                                                                                                      | false    |
| length      | number            | The message's length (bytes count)                                                                                                                                                             | true     |
| sender      | string            | The message's sender name                                                                                                                                                                      | false    |
| signals     | map[string]Signal | A map containing the message's signals, with the signal name as key                                                                                                                            | true     |
| attributes  | map[string]any    | A map with key the attribute name and a value to assign as map's value. The value must be an int if the attribute is of type int, string for type string, an enum value (string) for type enum | false    |

### Signal

| field       | type              | description                                                                                                                                                                                    | required                  | default |
| ----------- | ----------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------- | ------- |
| start_bit   | number            | The signal's start bit                                                                                                                                                                         | true                      |
| size        | number            | The signal's size (bits count)                                                                                                                                                                 | true                      |
| description | string            | The signal's description                                                                                                                                                                       | false                     |
| endianness  | little \| big     | The signal's byte order                                                                                                                                                                        | false                     | little  |
| signed      | boolean           | The signal's value type                                                                                                                                                                        | false                     | false   |
| receivers   | string[]          | The signal's receivers list                                                                                                                                                                    | false                     |
| scale       | number            | The signal's scale                                                                                                                                                                             | false                     | 1       |
| offset      | number            | The signal's offset                                                                                                                                                                            | false                     | 0       |
| min         | number            | The signal's minimum value                                                                                                                                                                     | false                     | 0       |
| max         | number            | The signal's maximum value                                                                                                                                                                     | true                      |
| enum        | map[string]number | A map with key the _uman readable_ name corrisponding to a signal's vlue                                                                                                                       | false                     |
| mux_group   | map[string]Signal | A map with key the name of a multiplexed signal and a Signal as value. If set, the signal becomes a multiplexor                                                                                | false                     |
| mux_switch  | number            | The value a multiplexor signal as to be in order to map to the multiplexed signal                                                                                                              | Only if part of mux_group |
| attributes  | map[string]any    | A map with key the attribute name and a value to assign as map's value. The value must be an int if the attribute is of type int, string for type string, an enum value (string) for type enum | false                     |
