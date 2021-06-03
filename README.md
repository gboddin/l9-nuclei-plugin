# L9 Nuclei plugin

This L9 plugin re-implements a limited subset of [Nuclei](https://github.com/projectdiscovery/nuclei), 
[ProjectDiscovery](https://github.com/projectdiscovery)'s awesome network scanner.

This allows for l9explore to stick to deep-protocol inspections while taking advantage of well maintained templates for
web application scanning.

## Features

- Uses upstream tag fields from `l9events` to match against nuclei template tags (`wordpress`,`php`)

## POC

This is currently a proof-of-concept and design may change.

There's a [pre-release](https://github.com/LeakIX/l9explore/releases) version of l9explore including this plugin.

## Usage

```sh
NUCLEI_TEMPLATES=/home/user/nuclei-templates ./l9explore service --debug
```