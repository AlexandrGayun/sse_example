***Prerequisites:***
1. Create a .env file based on the .env.example(.env file is necessary)
2. Provide a needed values to .env(HTTP_API_PORT is optional, 8080 by default, OPENAI_API_KEY is mandatory)

***How to run:***
1. `make up` will run the app in the docker
2. but it's also possible to build and run on your own

***Notes:***</br>
The app is designed for single user usage (in case of few /chat connections the resulting order in response is not guaranteed)</br>
The app supports /chat reconnect. All chat messages will await consumption by at least one sse client

***How to test:***
1. Connect with any client supporting websocket connections(Postman for example) to /chat endpoint
2. For SSE test you could use curl with -N key:</br>
`curl -N localhost:8080/sse` for example
