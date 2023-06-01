# Language Features
### Author: Creative Commons Zero
### Date: 2023-01-24

# Golang choice

Golang was chosen as a language that is like plain C.
It does not require the fancy extensions of Bjarne Stroustrup.
Those align really well with Physics, or Enterprise modelling projects.
This is infrastructure code and a simple language suits better.

# Golang features

We definitely advise against using any go.mod features.
We suggest copying each dependency and reviewing changes instead.
Most golang dependencies are simple, so it is less of an issue.
However, we use the http library that allows plugins, and we want to avoid any plugins redirecting sensitive calls.

We also do not generate code for simplicity.
It reduces engineering time.

We use a salt and activation key in metadata/data.go that may need to be refreshed occasionally.
Ideally they need to be unique in every production cluster.

We do not use fancy features of golang like generics.
This makes the code more portable to older versions or to dolang, our own distribution.

We do not use YAML, JSON, XML.
The reason is that why would you learn and use another language in the same tem.
Everybody should understand the whole solution.
All configuration is embedded into the sources.
This is why we advise against using Helm charts as well. It is just simpler.

We do not precompile but compile at startup due to golang being very good at compilation.

We do not handle exceptions and errors in many cases.
The consideration is that you make your infrastructure work first.
Handling edge cases has lower marginal engineering return.
A simple empty result represents an error that should never happen.
Also, sporadic error handling is an extra effort to test and check during security reviews.

## License

```
This document is Licensed under Creative Commons CC0.
To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
to this document to the public domain worldwide.
This document is distributed without any warranty.
You should have received a copy of the CC0 Public Domain Dedication along with this document.
If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.
```
