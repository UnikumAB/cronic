# cronic
A golang take on the cronic script https://habilis.net/cronic/ .

It will help you with those pesty cron mails that you don't want to see in case of success.
If the command it executes has an exit code other than 0 it prints the output split into stdout, stderr and trace output (if there is any).

## Usage
`LOGFILE_NAME=~/cronic.log cronic echo OK`

If you do not specify the environment variable `LOGFILE_NAME` then it will try to log to `/var/log/cronic.log`

