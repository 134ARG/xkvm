use serde::{Deserialize, Serialize};
use std::fs;
use std::path::PathBuf;
use tauri::Manager;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Connection {
    pub id: String,
    pub name: String,
    pub url: String,
    pub is_default: bool,
    pub last_connected: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConfigSettings {
    pub auto_connect: bool,
    pub remember_last_connection: bool,
}

impl Default for ConfigSettings {
    fn default() -> Self {
        Self {
            auto_connect: true,
            remember_last_connection: true,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AppConfig {
    pub version: String,
    pub connections: Vec<Connection>,
    pub settings: ConfigSettings,
}

impl Default for AppConfig {
    fn default() -> Self {
        Self {
            version: "1.0".to_string(),
            connections: Vec::new(),
            settings: ConfigSettings::default(),
        }
    }
}

fn get_config_dir(app_handle: tauri::AppHandle) -> Result<PathBuf, String> {
    let config_dir = app_handle
        .path()
        .app_data_dir()
        .map_err(|e| format!("Failed to get app data directory: {}", e))?;
    
    if !config_dir.exists() {
        fs::create_dir_all(&config_dir)
            .map_err(|e| format!("Failed to create config directory: {}", e))?;
    }
    
    Ok(config_dir)
}

fn get_config_path(app_handle: tauri::AppHandle) -> Result<PathBuf, String> {
    Ok(get_config_dir(app_handle)?.join("config.json"))
}

fn load_config_from_file(app_handle: tauri::AppHandle) -> Result<AppConfig, String> {
    let config_path = get_config_path(app_handle)?;
    
    if !config_path.exists() {
        return Ok(AppConfig::default());
    }
    
    let content = fs::read_to_string(&config_path)
        .map_err(|e| format!("Failed to read config file: {}", e))?;
    
    let config: AppConfig = serde_json::from_str(&content)
        .map_err(|e| format!("Failed to parse config file: {}", e))?;
    
    Ok(config)
}

fn save_config_to_file(app_handle: tauri::AppHandle, config: &AppConfig) -> Result<(), String> {
    let config_path = get_config_path(app_handle)?;
    
    let content = serde_json::to_string_pretty(config)
        .map_err(|e| format!("Failed to serialize config: {}", e))?;
    
    fs::write(&config_path, content)
        .map_err(|e| format!("Failed to write config file: {}", e))?;
    
    Ok(())
}

pub fn init_config(app_handle: tauri::AppHandle) -> Result<(), String> {
    let config = load_config_from_file(app_handle.clone())?;
    save_config_to_file(app_handle, &config)?;
    Ok(())
}

#[tauri::command]
pub fn get_config(app_handle: tauri::AppHandle) -> Result<AppConfig, String> {
    load_config_from_file(app_handle)
}

#[tauri::command]
pub fn save_config(app_handle: tauri::AppHandle, config: AppConfig) -> Result<(), String> {
    save_config_to_file(app_handle, &config)
}

#[tauri::command]
pub fn add_connection(app_handle: tauri::AppHandle, name: String, url: String) -> Result<Connection, String> {
    let mut config = load_config_from_file(app_handle.clone())?;
    
    // Generate unique ID
    let id = format!("conn_{}", chrono::Utc::now().timestamp());
    
    // If this is the first connection, make it default
    let is_default = config.connections.is_empty();
    
    let connection = Connection {
        id: id.clone(),
        name,
        url,
        is_default,
        last_connected: None,
    };
    
    config.connections.push(connection.clone());
    save_config_to_file(app_handle, &config)?;
    
    Ok(connection)
}

#[tauri::command]
pub fn update_connection(
    app_handle: tauri::AppHandle,
    id: String,
    name: String,
    url: String,
) -> Result<Connection, String> {
    let mut config = load_config_from_file(app_handle.clone())?;

    let connection = config
        .connections
        .iter_mut()
        .find(|c| c.id == id)
        .ok_or("Connection not found")?;

    connection.name = name;
    connection.url = url;

    let updated = connection.clone();
    save_config_to_file(app_handle, &config)?;

    Ok(updated)
}

#[tauri::command]
pub async fn test_connection(url: String) -> Result<(), String> {
    let base = url.trim_end_matches('/');
    let target = format!("{}/device/status", base);

    let client = tauri_plugin_http::reqwest::Client::builder()
        .timeout(std::time::Duration::from_secs(5))
        .build()
        .map_err(|e| format!("Failed to build HTTP client: {}", e))?;

    client
        .get(&target)
        .send()
        .await
        .map_err(|e| format!("Could not reach backend: {}", e))?;

    // Any HTTP response means the backend is reachable.
    Ok(())
}

#[tauri::command]
pub fn remove_connection(app_handle: tauri::AppHandle, id: String) -> Result<(), String> {
    let mut config = load_config_from_file(app_handle.clone())?;
    
    let was_default = config.connections.iter()
        .find(|c| c.id == id)
        .map(|c| c.is_default)
        .unwrap_or(false);
    
    config.connections.retain(|c| c.id != id);
    
    // If we removed the default connection, make the first one default
    if was_default && !config.connections.is_empty() {
        config.connections[0].is_default = true;
    }
    
    save_config_to_file(app_handle, &config)?;
    Ok(())
}

#[tauri::command]
pub fn set_default_connection(app_handle: tauri::AppHandle, id: String) -> Result<(), String> {
    let mut config = load_config_from_file(app_handle.clone())?;
    
    // Unset all defaults
    for conn in &mut config.connections {
        conn.is_default = false;
    }
    
    // Set the new default
    let found = config.connections.iter_mut()
        .find(|c| c.id == id)
        .ok_or("Connection not found")?;
    
    found.is_default = true;
    
    save_config_to_file(app_handle, &config)?;
    Ok(())
}

#[tauri::command]
pub fn update_last_connected(app_handle: tauri::AppHandle, id: String) -> Result<(), String> {
    let mut config = load_config_from_file(app_handle.clone())?;
    
    let connection = config.connections.iter_mut()
        .find(|c| c.id == id)
        .ok_or("Connection not found")?;
    
    connection.last_connected = Some(chrono::Utc::now().to_rfc3339());
    
    save_config_to_file(app_handle, &config)?;
    Ok(())
}
