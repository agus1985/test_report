# Report GitHub Pull Requests


## Docker


##### Build the docker image
```sh
docker build -t <image_name>:<tag> .
```


##### To run the report set the followint mandatory environment variables
```sh
export apiToken=<token>
export apiUrl="https://api.github.com"
export emailRecipientsReport="<emailaddr1,emailaddr2,...emailaddrn>"
export githubUser="zalando"
export githubRepo="patroni"
export emailSender="report_sender@example.com"
```
And then run the docker container
```sh
docker run -it \
-e apiUrl=$apiUrl \
-e apiToken=$apiToken \
-e emailRecipientsReport=$emailRecipientsReport \
-e githubUser=$githubUser \
-e githubRepo=$githubRepo \
-e emailSender=$emailSender \
```

## OR
##### Use provided scripts

> Note: `.vars.sh ` for setting envrionment variables
> Note: `test.sh ` for running test

##### Run
```sh
./test.sh
```

> Optional variables: 
`dryRun ` set to false will send email over smtp to defined ${emailRecipientsReport}
If `dryRun ` set to false its Mandatory to define smtp  related configuration
`smtpUser` user credentials for smtp server
`smtpPassword` password credentials for smtp server
`smtpServer` smtp hostname
`smtpPort` smtp port use TLS based port