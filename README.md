## Javaman

Switch between java and maven versions on Windows.

## Requirements

- Python 3.x.x

## How to install

- Clone this repo to any folder
- Add the project folder to your path
- Update ``jdks.properties`` and ``mvn.properties`` with your java/maven installations

#### ``jdks.properties`` example

```
23=C:\Users\mathe\.jdks\openjdk-23.0.1
11=C:\Users\mathe\.jdks\corretto-11.0.26
```

#### ``mvn.properties`` example

```
3.9.10=C:\Users\mathe\.jdks\openjdk-23.0.1
3.8.8=C:\Users\mathe\.jdks\corretto-11.0.26
```

### Examples of usage:

Set java version to 11: ``jm use-jdk 11``

Set maven version to 3.9.9: ``jm use-mvn 3.9.9``

List installed jdks: ``jm list jdk``

List installed mvns: ``jm list mvn``
