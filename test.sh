docker build -t github_report .
. ./.vars.sh
docker run -it \
-e apiUrl=$apiUrl \
-e apiToken=$apiToken \
-e emailRecipientsReport=$emailRecipientsReport \
-e githubUser=$githubUser \
-e githubRepo=$githubRepo \
-e emailSender=$emailSender \
-e dryRun=$dryRun \
github_report 