## Javaman

Switch between java and maven versions.

## Setup


For the tool to work, it is necessary to configure two files pointing to the Java and/or Maven installations. For Java you need the ``jdks.properties`` file, for maven ``mvns.properties``. The files must be in %USERPROFILE%/javaman.


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

### Example of usage:

Set java version to 11: ``javaman --jdk 11``

Set maven version to 3.9.9: ``javaman --mvn 3.9.9``