# Security considerations
### Author: Author
### Date: 2023-05-19

We collected the security considerations and design decisions regarding the project.

## Threat Model

Computer systems run in a different security space than before.
Geopolitical risks have changed and the situation does not seem to get better as of today.
Even the smallest warehousing software needs to deal with the best, well funded government supported attackers, and their support base.

We designed a solution with the concept of "steel containers", a completely zero trust model.
The basic concept is that accounts, passwords and roles used to protect against insider threat.
Nowadays, the threat are the devices that you have around yourself, e.g. [Snake Malware](https://www.cbsnews.com/news/fbi-takes-down-20-year-old-russian-malware-network/)
This means that each application has to protect itself from other applications of the same user and machine.

## Architecture

This is an application level project.
It relies on the cloud provider and the language platform to provide standard security facilities.
We do not deal with TLS directly. We rely on the load balancers of the cloud providers.
This makes our stack relatively free of easy targets.


## Language platform

Go has an extensive built-in http library increasing overall risk.
The official built in http server of Go allows test hooks.
Care should be taken with third party libraries,
so that they do not tamper with the main security path.

We can use it as long as there are no external modules that can add additional debug listeners etc.
The project code is smaller this way increasing the security of the project itself.
This is why the project prohibits the use of external go.mod modules.
We still use official Go binaries.

However, we do not use any external libraries to reduce this risk.
The problem with external modules can be described with a stochastic model.
These modules must be checked a hundred times, if they have a change hundred times a year. 
Using fifty modules would add an unnecessary burden of five thousand reviews to the development team.
There is not many external libraries that justify such an expense.
Generative copilots can generate many algorithms once.

The project code is as small as possible following Go design standards.
We chose this language especially for this reason compared to Python, C#, or Java.
This will help to reduce risk by lowering ramp up time.
It requires a smaller engineering team required.
It simplifies audits.

## Cryptography by cloud

We use TLS provided by the cloud provider.
We could do it ourselves, but this matches the design principle.
Institutional applications should consider adding additional layers of security.
Our company can provide symmetric private key encryption solutions upon request (www.schmied.us).
Geopolitical risk may cause some certificate servers close to the root compromised,
generating fake certificates for your sites.
Malicious actors can run cloned copies stealing money and data this way.

There is no reason not to trust your cloud provider, if they own the physical location already.
A stronger application layer TLS would be unnecessary, a less strong one would be risky.

## Tokens

We use apikey instead of public key and private key tokens for authentication.
We avoid the need to keep browser cookies on the client side eliminating
any regulatory requirements to opt in and accept agreements.

This can actually help advertising sites to show a clear picture right from the beginning.
Sites can lose customers due to the burden of "Accept Cookies" buttons just for a few cents of ad space.
These buttons are distraction.

Api keys are equivalent to private keys being a sufficiently long cryptographic random numbers.
They are stronger than public key encryption not adding additional metadata to the keys like domains.
A stolen key is practically useless without the container that uses it.
Public key encryption was designed for many users of a site, but who needs it for peer-to-peer communication?

## Browser environment

Api keys are technically kept in browser history instead of Authentication tokens in cookies.
Browser history may be backed up to cloud storage.
This is still personal. It may be accessed as document referrer in Javascript.
It allows on the other hand to have messages sent just with a URL just like what Zoom does.

Institutional usage must have organizational policies to keep browser history secret, or do not collect history.
Also, we do not allow the use of external Javascript libraries for this reason.

We do reasonable effort to protect messages by not keeping them longer than a configurable time, typically 148 hours.
Even if an apikey is stolen, the recipient reads messages by this time.
Attackers cannot access old messages.

## The Maths behind

**Example.** A pin code consists of five digits.
The authentication code delays three seconds.
An attacker without a hint needs a brute force attack of more than three days to try them all.
This even makes it risky to open a deposit box throughout a long weekend.

Obviously it does not protect against attackers, who have a fifteen minutes access every weekend for a prolonged period.
However, such attackers would plant a camera to record the password anyway. We rely on our cloud providers.

## Brute force attacks

We use a similar approach with our Api keys.
We use 64 of the 26 latin letters and waiting 15 ms with the pass and fail results.
This makes it impossible to guess the valid Api keys.

Non-legitimate clients need to carry out parallel trials to bypass the delay logic.
We add a mutex for this reason.

## Advanced persistent threats.

Sophisticated attackers will try to attach the underlying infrastructure.
They may change the code or tamper with random generation.
The [Snake Malware](https://www.cbsnews.com/news/fbi-takes-down-20-year-old-russian-malware-network/) is a good example.

This infrastructure topic not covered by this application project.
Various solutions like obtaining random numbers from external servers and XOR them practically makes such attacks useless.
They force attackers to focus on a single final point of usage, where they can be caught.

## Denial of Service

Denial of Service attacks are resolved by waiting for a reasonable time (15 ms) at Api key approval.
Legitimate clients will get the full CPU after the waiting time.
We simply scale the cluster, so that it has a matching CPU power to the bandwidth.

New TCP and HTTP sessions will likely be prioritized lower over existing TCP sessions of IP addresses by internet routing infrastructure. Again, we rely on infrastructure. A good load balancer choice can prevent these without application changes.

We may use multiple server containers, but we will route traffic to the same container sharded by the api key.

## Browser choice

There has not been any reason found, why tokens and authentication cookies would be more secure than api keys of page URLs in a general browser environment.
Specific hardware platforms and browsers may provide environments protected better with hardware encryption, process isolation, etc.
but we target the general engineering audience.
Use private browser contexts, whenever it is needed.

Also, relying on hardware encryption requires expensive expert staff to verify the case of faulty TPM chip lots by target countries, etc.

It cannot be broken with traditional or quantum computers in reasonable amount of time.
The authentication and zero trust authorization of 64 letter tokens should be strong enough to protect against attacks by quantum computers.
We call this the El Alamein effect. The application level solution can be made up to Quantum Grade Security provided that the channel encryption is symmetric.
