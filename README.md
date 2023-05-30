Receipt Processor

This is a takehome test form fetch-rewards. 

Assumptions: 

- time (24:00) is not valid and would instead be 00:00
- There are some point calculations based on the "trimmed length" of the short description. I assumed this to be the length without the preceeding and following spaces.
- Allowed for future dates
- Considered odd dates to be dates with an odd day.

Requirements to run:

- Must execute the following command for a required dependecny: go get github.com/google/uuid

How to run:

- Simply run go tun main.go in terminal
- make get and post requests to localhost:8080 and the appropriate endpoints as would normally example using curl or postman.