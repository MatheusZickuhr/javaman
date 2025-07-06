use std::fs::File;
use std::io::BufRead;
use std::path::{Path, PathBuf};
use std::{env, io};
use winreg::{RegKey, enums::*};

struct Installation {
    home_path: String,
    bin_path: String,
    version: String,
}

struct ParsedArgs {
    jdk_version: Option<String>,
    mvn_version: Option<String>,
}

fn main() {
    let args: Vec<String> = env::args().collect();

    let parsed_args: ParsedArgs = parse_args(args);

    let javaman_folder = get_javaman_folder();
    
    if parsed_args.jdk_version.is_some() {
        let jdk_version = parsed_args.jdk_version.unwrap();
        let prop_file_path = get_prop_file_path(&javaman_folder, "jdks.properties");
        
        update_version(jdk_version, "JAVA_HOME", prop_file_path.as_str());
    }

    if parsed_args.mvn_version.is_some() {
        let mvn_version = parsed_args.mvn_version.unwrap();
        let prop_file_path = get_prop_file_path(&javaman_folder, "mvns.properties");
        
        update_version(mvn_version, "MAVEN_HOME", prop_file_path.as_str());
    }
}

fn get_javaman_folder() -> PathBuf {
    let mut javaman_folder = get_user_folder();
    javaman_folder.push("javaman");
    javaman_folder
}

fn get_prop_file_path(javaman_folder: &PathBuf, prop_file_name: &str) -> String {
    let mut javaman_folder_clone = javaman_folder.clone();
    javaman_folder_clone.push(prop_file_name);
    
    let prop_file = javaman_folder_clone
        .as_path()
        .to_str()
        .unwrap();
    
    prop_file.to_string()
}

fn get_user_folder() -> PathBuf {
    let user_folder = env::var("USERPROFILE").unwrap();
    PathBuf::from(user_folder)
}

fn update_version(version: String, home_env_variable: &str, config_file_path: &str) {
    let installations = read_installations_from_config_file(config_file_path);

    let selected_installation = installations
        .iter()
        .find(|&installation| installation.version == version)
        .unwrap();

    let path_env_variable = "Path";

    let system_path = read_env_variable(path_env_variable);
    let mut new_path = remove_installations_from_path(system_path, &installations);
    new_path = add_new_installation_to_path(new_path, selected_installation);

    set_env_variable(path_env_variable, new_path.as_str());
    set_env_variable(home_env_variable, selected_installation.home_path.as_str());
}

fn parse_args(args: Vec<String>) -> ParsedArgs {
    let mut jdk_version_option = None;
    let mut mvn_version_option = None;

    for (i, arg) in args.iter().enumerate() {
        if arg == "--jdk" {
            jdk_version_option = Some(args[i + 1].clone());
        } else if arg == "--mvn" {
            mvn_version_option = Some(args[i + 1].clone());
        }
    }

    ParsedArgs {
        jdk_version: jdk_version_option,
        mvn_version: mvn_version_option,
    }
}

fn add_new_installation_to_path(
    mut system_path: String,
    selected_installation: &Installation,
) -> String {
    system_path.push_str(";");
    system_path.push_str(selected_installation.bin_path.as_str());
    system_path
}

fn read_installations_from_config_file(config_file_name: &str) -> Vec<Installation> {
    let file_path = Path::new(config_file_name);
    let file = File::open(&file_path).expect("Não foi possível abrir o arquivo");
    let buf_reader = io::BufReader::new(file);

    let mut installations: Vec<Installation> = Vec::new();

    for line_result in buf_reader.lines() {
        let line = line_result.expect("Erro ao ler linha");
        let parts = line.split("=").collect::<Vec<&str>>();
        let key = parts[0].to_string();
        let value = parts[1].to_string();

        let mut bin_folder = value.clone();
        bin_folder.push_str("\\bin");

        installations.push(Installation {
            version: key,
            home_path: value,
            bin_path: bin_folder,
        });
    }

    installations
}

fn remove_installations_from_path(
    system_path: String,
    installations: &Vec<Installation>,
) -> String {
    let bin_folders = system_path.split(";").collect::<Vec<&str>>();

    let mut new_path: String = String::new();

    for (i, bin_folder) in bin_folders.iter().enumerate() {
        if is_bin_folder_on_installations(bin_folder, &installations) {
            continue;
        }

        new_path.push_str(bin_folder);

        if i < bin_folders.len() - 1 {
            new_path.push_str(";");
        }
    }

    new_path
}

fn is_bin_folder_on_installations(bin_folder: &str, installations: &Vec<Installation>) -> bool {
    for installation in installations {
        if bin_folder.contains(installation.home_path.as_str()) {
            return true;
        }
    }
    false
}

fn read_env_variable(name: &str) -> String {
    let hkcu = RegKey::predef(HKEY_CURRENT_USER);
    let env = hkcu.open_subkey("Environment").unwrap();

    let value: String = env.get_value(name).unwrap();
    // println!("Valor da variável: {}", value);

    value
}

fn set_env_variable(name: &str, value: &str) {
    // variavel de ambiente
    let hkcu = RegKey::predef(HKEY_CURRENT_USER);
    let (env, _) = hkcu.create_subkey("Environment").unwrap();

    env.set_value(name, &value).unwrap();
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_remove_installation_from_path() {
        let system_path =
            "C:\\Users\\mathe\\AppData\\Local\\Programs\\Python\\Python313\\Scripts\\;\
        C:\\Users\\mathe\\AppData\\Local\\Programs\\Python\\Python313\\;\
        C:\\Users\\mathe\\AppData\\Local\\Programs\\Python\\Launcher\\;\
        C:\\Users\\mathe\\.cargo\\bin;\
        C:\\Users\\mathe\\AppData\\Local\\Microsoft\\WindowsApps;\
        C:\\Users\\mathe\\AppData\\Local\\Programs\\Microsoft VS Code\\bin;\
        C:\\Users\\mathe\\AppData\\Roaming\\npm;\
        C:\\Users\\mathe\\.jdks\\openjdk-23.0.1\\bin;\
        C:\\Program Files\\Microsoft Visual Studio\\2022\\Community\\MSBuild\\Current\\Bin;\
        C:\\Users\\mathe\\AppData\\Local\\gitkraken\\bin;\
        C:\\Program Files\\LLVM\\bin"
                .to_string();

        let mut installations: Vec<Installation> = Vec::new();
        installations.push(Installation {
            bin_path: "C:\\Users\\mathe\\.jdks\\corretto-11.0.26\\bin".to_string(),
            home_path: "C:\\Users\\mathe\\.jdks\\corretto-11.0.26".to_string(),
            version: "11".to_string(),
        });

        installations.push(Installation {
            bin_path: "C:\\Users\\mathe\\.jdks\\openjdk-23.0.1\\bin".to_string(),
            home_path: "C:\\Users\\mathe\\.jdks\\openjdk-23.0.1".to_string(),
            version: "23".to_string(),
        });

        let expected = "C:\\Users\\mathe\\AppData\\Local\\Programs\\Python\\Python313\\Scripts\\;\
        C:\\Users\\mathe\\AppData\\Local\\Programs\\Python\\Python313\\;\
        C:\\Users\\mathe\\AppData\\Local\\Programs\\Python\\Launcher\\;\
        C:\\Users\\mathe\\.cargo\\bin;\
        C:\\Users\\mathe\\AppData\\Local\\Microsoft\\WindowsApps;\
        C:\\Users\\mathe\\AppData\\Local\\Programs\\Microsoft VS Code\\bin;\
        C:\\Users\\mathe\\AppData\\Roaming\\npm;\
        C:\\Program Files\\Microsoft Visual Studio\\2022\\Community\\MSBuild\\Current\\Bin;\
        C:\\Users\\mathe\\AppData\\Local\\gitkraken\\bin;\
        C:\\Program Files\\LLVM\\bin"
            .to_string();

        let actual = remove_installations_from_path(system_path, &installations);

        assert_eq!(actual, expected);
    }

    #[test]
    fn test_add_new_installation_to_path() {
        let system_path =
            "C:\\Users\\mathe\\AppData\\Local\\Programs\\Python\\Python313\\Scripts\\;\
        C:\\Users\\mathe\\AppData\\Local\\Programs\\Python\\Python313\\;\
        C:\\Users\\mathe\\AppData\\Local\\Programs\\Python\\Launcher\\;\
        C:\\Users\\mathe\\.cargo\\bin;\
        C:\\Users\\mathe\\AppData\\Local\\Microsoft\\WindowsApps;\
        C:\\Users\\mathe\\AppData\\Local\\Programs\\Microsoft VS Code\\bin;\
        C:\\Users\\mathe\\AppData\\Roaming\\npm;\
        C:\\Program Files\\Microsoft Visual Studio\\2022\\Community\\MSBuild\\Current\\Bin;\
        C:\\Users\\mathe\\AppData\\Local\\gitkraken\\bin;\
        C:\\Program Files\\LLVM\\bin"
                .to_string();

        let selected_installation = Installation {
            bin_path: "C:\\Users\\mathe\\.jdks\\corretto-11.0.26\\bin".to_string(),
            home_path: "C:\\Users\\mathe\\.jdks\\corretto-11.0.26".to_string(),
            version: "11".to_string(),
        };

        let expected = "C:\\Users\\mathe\\AppData\\Local\\Programs\\Python\\Python313\\Scripts\\;\
        C:\\Users\\mathe\\AppData\\Local\\Programs\\Python\\Python313\\;\
        C:\\Users\\mathe\\AppData\\Local\\Programs\\Python\\Launcher\\;\
        C:\\Users\\mathe\\.cargo\\bin;\
        C:\\Users\\mathe\\AppData\\Local\\Microsoft\\WindowsApps;\
        C:\\Users\\mathe\\AppData\\Local\\Programs\\Microsoft VS Code\\bin;\
        C:\\Users\\mathe\\AppData\\Roaming\\npm;\
        C:\\Program Files\\Microsoft Visual Studio\\2022\\Community\\MSBuild\\Current\\Bin;\
        C:\\Users\\mathe\\AppData\\Local\\gitkraken\\bin;\
        C:\\Program Files\\LLVM\\bin;\
        C:\\Users\\mathe\\.jdks\\corretto-11.0.26\\bin"
            .to_string();

        let actual = add_new_installation_to_path(system_path, &selected_installation);

        assert_eq!(actual, expected);
    }
}
