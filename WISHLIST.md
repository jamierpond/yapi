# YAPI Wishlist

Issues encountered while writing MIDI clip note tests against the Ableton Live Remote Script.

## 1. Chain variable syntax not documented clearly

YAPI uses `${step.result.field}` (curly braces required), but the old docs and some examples show `$step.result.field` (bare dollar sign). The bare form is silently passed as a literal string rather than substituted, which makes debugging painful — the Remote Script receives the string `"$create_track.result.index"` instead of `3`.

**Wish:** Either support both forms, or emit a clear warning when a string matching `$word.word` is found without braces.

## 2. No way to inspect resolved variable values

When a chain step fails, the output shows the raw response but not the resolved parameter values that were sent. If variable substitution silently fails (e.g., wrong syntax), you can't tell from the output.

**Wish:** Show the resolved request body in verbose/debug mode so you can verify what was actually sent over the wire.

## 3. No way to print step results mid-chain for debugging

When a chain step fails, you get the response for the failing step but not intermediate steps (unless you scroll through the full output). Being able to mark a step as `debug: true` to print its full response would help.

**Wish:** `debug: true` on chain steps to always print the full response body, or a `--verbose` flag that prints all step responses.
