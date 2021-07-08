# L9 Nuclei plugin

This L9 plugin re-implements a limited [Nuclei](https://github.com/projectdiscovery/nuclei) template parser and runner.

Checkout [ProjectDiscovery](https://github.com/projectdiscovery)'s awesome network tools for more information.

This allows for `l9explore` to stick to deep-protocol inspections while taking advantage of well maintained templates for
web application scanning.

## Features

- Uses upstream tag fields from `l9events` to match against nuclei template tags (`wordpress`,`php`)

## POC

This is currently a proof-of-concept and design may change.

There's a [pre-release](https://github.com/LeakIX/l9explore/releases) version of l9explore including this plugin.

## Settings

```sh
# Nuclei template directory location :
export NUCLEI_TEMPLATES=/home/user/nuclei-templates
# Tags to ALWAYS run during scans :
export NUCLEI_DEFAULT_TAGS=exposure
# List of template IDs to disable :
export NUCLEI_DISABLED_TEMPLATES=git-config,CVE-2017-5487,default-nginx-page
```

## Usage

```sh
NUCLEI_TEMPLATES=/home/user/nuclei-templates ./l9explore service --debug
```
