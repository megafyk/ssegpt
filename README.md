# SSEGPT

Simple golang server send event with chatgpt.

## Usage

```docker
docker run -d -p 8080:8080 \
-e OPENAI_API_URL=https://api.openai.com/v1/completions
-e OPENAI_API_KEY=<your openai api key> \
megafyk/ssegpt:latest
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

[MIT](https://choosealicense.com/licenses/mit/)