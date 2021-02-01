# sw.golibs

## How to contribute

You can find a lot of information in the [contributing documentation](.github/CONTRIBUTING.md)

## log
* RotateWriter: writer to handle log rotation.
* CollapsingWriter: pack duplicate log messages into a single entry followed by xN with N the number of occurrences.

## windows
* MakeProcessKillItsSubProcess(): ensure sub process are killed when parent is killed.

## config
* Config: parses the given configuration file and updates it with the currently defined flags. The file gets created if it doesn't exist.

## ts
* Tools to read/parse/generate TS files
