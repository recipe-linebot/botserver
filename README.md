# recipe-linebot/botserver

_This documentaion is still under construction._ :construction_worker:

_botserver_ is a component of _recipe-linebot_.
It behaves as LINE bot, that is accepts callback events and send back response to client.
Because _recipe-linebot_ is a service to suggest recipes by names of food stuff, so _botserver_ accepts the message event includes the names of food stuff and sends back suggested recipes to friend user as client.

## Development

- Use Git as VCS.
- The repository is hosted by GitHub.
- The workflow is based on [GitHub Flow](https://guides.github.com/introduction/flow/).
- The commit message style is recommended: `{emoji} {space} {message}`.
  - `emoji` should be specified by `:name:`.
  - A kind of emoji should be chosen on [Gitmoji](https://gitmoji.carloscuesta.me/).
  - `message` should start with a present form of verb and not end with a period.
