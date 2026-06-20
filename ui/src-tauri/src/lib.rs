mod config;

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
  tauri::Builder::default()
    .plugin(tauri_plugin_http::init())
    .plugin(tauri_plugin_fs::init())
    .invoke_handler(tauri::generate_handler![
      config::get_config,
      config::save_config,
      config::add_connection,
      config::update_connection,
      config::test_connection,
      config::remove_connection,
      config::set_default_connection,
      config::update_last_connected,
    ])
    .setup(|app| {
      if cfg!(debug_assertions) {
        app.handle().plugin(
          tauri_plugin_log::Builder::default()
            .level(log::LevelFilter::Info)
            .build(),
        )?;
      }
      
      // Initialize config on first run
      if let Err(e) = config::init_config(app.handle().clone()) {
        log::error!("Failed to initialize config: {}", e);
      }
      
      Ok(())
    })
    .run(tauri::generate_context!())
    .expect("error while running tauri application");
}
