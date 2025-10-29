# Javaman

> Manage and switch between multiple JDK and Maven versions on Windows.

## Features

- Install and uninstall specific JDK and Maven versions
- Switch active JDK or Maven version (updates environment variables)
- List all installed JDKs and Maven versions
- Simple CLI interface

## Installation

- Download from releases page [here](https://github.com/MatheusZickuhr/javaman/releases) 
- Add the application folder to your user's path environment variable

## Usage

### Install JDK or Maven

```powershell
javaman install-jdk <version>
javaman install-mvn <version>
```

### Switch active version

```powershell
javaman use-jdk <version>
javaman use-mvn <version>
```

### List installed versions

```powershell
javaman list jdk
javaman list mvn
```

### Uninstall JDK or Maven

```powershell
javaman uninstall-jdk <version>
javaman uninstall-mvn <version>
```

## Example

```powershell
javaman install-jdk 11
javaman use-jdk 11
javaman install-mvn 3.9.9
javaman use-mvn 3.9.9
javaman list jdk
javaman list mvn
javaman uninstall-jdk 11
javaman uninstall-mvn 3.9.9
```

## How it works

- Installs JDK/Maven to a managed directory
- Updates `JAVA_HOME`, `MAVEN_HOME`, and system `Path` for the selected version
- Stores configuration in `javaman.json`
