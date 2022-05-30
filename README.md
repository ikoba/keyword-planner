# keyword-planner

keyword-planner is a command line tool to automatically retrieve keyword CSV files from Google Keyword Planner using the Google Chrome.

## Installation

`go install github.com/ikoba/keyword-planner/cmd/kp@latest`

## Usage

1. Kill all Google Chrome proccesses.
2. Launch Google Chrome with the option `--remote-debugging-port=9222`.
   If you are using macOS, the command may be as follows.

```
/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --remote-debugging-port=9222
```

To find out Google Chrome installation path, type the following in the address bar.

```
chrome://version/
```

3. Sign in to Google Ads.
4. Execute `kp` command in terminal.

```
kp -out OUTPUT_DIRECTORY KEYWORD1 KEYWORD2 ...
```

To see the options details, execute the following command.

```
kp -h
```

## License

MIT
