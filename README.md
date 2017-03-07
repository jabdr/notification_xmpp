# notification_xmpp
This is a nagios/icinga/naemon notification command.
## Build

```bash
# export GOPATH=$(pwd)
go get github.com/jabdr/notification_xmpp
```

## Howto

### Configuration file
We need a configuration file that stores the xmpp client settings:

```
username: openitcockpit@example.com
password: MYSUPERSECRETPASSWORD
starttls: true
ignorecertificate: false
html: false
```

### Template file
The command itself doesn't have a builtin message. You have to create your own message with the golang template engine and pass the parameters. Here is an example:

```
### ---
Hello {{.contactalias}},

{{.notificationtype}}

Host: {{.hostname}}
Host State: {{.hoststate}}
Host Output: {{.hostoutput}}
{{if .hostackauthor}}{{/*
*/}}Acknowledgement: {{.hostackauthor}}
Acknowledgement Message: {{.hostackcomment}}
{{end}}{{/*
*/}}
{{if .servicedesc}}
Service: {{.servicedesc}}
Service State: {{.servicestate}}
Service Output: {{.serviceoutput}}
{{if .serviceackauthor}}
Acknowledgement: {{.serviceackauthor}}
Acknowledgement Message: {{.serviceackcomment}}
{{end}}{{/*
*/}}{{end}}{{/*
*/}}
Greetings

Your openITCOCKPIT Monitoring System
--- ###
```

### Notification Command
Finally we can specify the command in the configuration.

#### Host
```
$USER1$/notification_xmpp "-configuration-file=/opt/openitc/nagios/etc/xmpp.yml" "-template-file=/opt/openitc/nagios/etc/xmpp.txt" "-target-username=$CONTACTEMAIL$" "contactalias=$CONTACTALIAS$" "hostname=$HOSTNAME$" "notificationtype=$NOTIFICATIONTYPE$" "hoststate=$HOSTSTATE$" "hostoutput=$HOSTOUTPUT$" "hostackauthor=$HOSTACKAUTHOR$" "hostackcomment=$HOSTACKCOMMENT$"
```

#### Service
```
$USER1$/notification_xmpp "-configuration-file=/opt/openitc/nagios/etc/xmpp.yml" "-template-file=/opt/openitc/nagios/etc/xmpp.txt" "-target-username=$CONTACTEMAIL$" "contactalias=$CONTACTALIAS$" "hostname=$HOSTNAME$" "notificationtype=$NOTIFICATIONTYPE$" "hoststate=$HOSTSTATE$" "hostoutput=$HOSTOUTPUT$" "servicedesc=$SERVICEDESC$" "servicestate=$SERVICESTATE$" "serviceoutput=$SERVICEOUTPUT$" "serviceackauthor=$SERVICEACKAUTHOR$" "serviceackcomment=$SERVICEACKCOMMENT$"
```
