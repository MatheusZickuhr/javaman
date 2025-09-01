import sys
import winreg

MVNS_PROPERTIES_FILE = "mvns.properties"
JDKS_PROPERTIES_FILE = "jdks.properties"


class Installation:
    def __init__(self, version, home_path, bin_path):
        self.version = version
        self.home_path = home_path
        self.bin_path = bin_path


class ParsedArgs:
    def __init__(self, jdk_version=None, mvn_version=None, list_command=None):
        self.jdk_version = jdk_version
        self.mvn_version = mvn_version
        self.list_command = list_command


def main():
    args = sys.argv[1:]
    parsed_args = parse_args(args)

    if parsed_args.jdk_version:
        update_version(parsed_args.jdk_version, "JAVA_HOME", JDKS_PROPERTIES_FILE)

    if parsed_args.mvn_version:
        update_version(parsed_args.mvn_version, "MAVEN_HOME", MVNS_PROPERTIES_FILE)

    if parsed_args.list_command:
        execute_list_command(parsed_args.list_command)


def execute_list_command(list_command):
    prop_file_path = None
    if list_command == 'jdk':
        prop_file_path = JDKS_PROPERTIES_FILE

    elif list_command == 'mvn':
        prop_file_path = MVNS_PROPERTIES_FILE

    else:
        raise Exception("Invalid list command")

    print(f"reading from {prop_file_path}:")

    prop_file = open(prop_file_path, "r")

    for line in prop_file:
        print(line.replace("\n", ""))


def get_prop_file_path(javaman_folder, prop_file_name):
    return str(javaman_folder / prop_file_name)


def update_version(version, home_env_variable, config_file_path):
    installations = read_installations_from_config_file(config_file_path)
    selected_installation = find_selected_installation(installations, version)

    if selected_installation is None:
        print("No installation found for version " + version + " in file " + config_file_path)
        return

    path_env_variable = "Path"
    system_path = read_env_variable(path_env_variable)
    new_path = remove_installations_from_path(system_path, installations)
    new_path = add_new_installation_to_path(new_path, selected_installation)

    set_env_variable(path_env_variable, new_path)
    set_env_variable(home_env_variable, selected_installation.home_path)

    print(f"Successfully updated to version {version}, restart your terminal to apply changes")


def find_selected_installation(installations, version):
    for installation in installations:
        if installation.version == version:
            return installation

    return None


def parse_args(args):
    jdk_version = None
    mvn_version = None
    list_command = None

    for i, arg in enumerate(args):
        if arg == "use-jdk":
            jdk_version = args[i + 1]
        elif arg == "use-mvn":
            mvn_version = args[i + 1]
        elif arg == 'list':
            list_command = args[i + 1]

    return ParsedArgs(jdk_version, mvn_version, list_command)


def add_new_installation_to_path(system_path, selected_installation):
    return system_path + ";" + selected_installation.bin_path


def read_installations_from_config_file(config_file_name):
    installations = []
    with open(config_file_name, "r", encoding="utf-8") as file:
        for line in file:
            line = line.strip()
            if "=" in line:
                key, value = line.split("=", 1)
                bin_folder = value + "\\bin"
                installations.append(Installation(key, value, bin_folder))
    return installations


def remove_installations_from_path(system_path, installations):
    bin_folders = system_path.split(";")
    new_path_parts = [
        folder for folder in bin_folders
        if not is_bin_folder_on_installations(folder, installations)
    ]
    return ";".join(new_path_parts)


def is_bin_folder_on_installations(bin_folder, installations):
    return any(installation.home_path in bin_folder for installation in installations)


def read_env_variable(name):
    with winreg.OpenKey(winreg.HKEY_CURRENT_USER, r"Environment") as key:
        value, _ = winreg.QueryValueEx(key, name)
        return value


def set_env_variable(name, value):
    with winreg.CreateKey(winreg.HKEY_CURRENT_USER, r"Environment") as key:
        winreg.SetValueEx(key, name, 0, winreg.REG_SZ, value)


if __name__ == "__main__":
    main()
