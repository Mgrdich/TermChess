//! TermChess binary entry point (port of Go `cmd/termchess/main.go`).

use std::io::{self, Write};
use std::process::ExitCode;
use std::time::Duration;

use updater::{
    compare_versions, current_goarch, current_goos, detect_install_method, get_binary_filename,
    get_go_install_message, uninstall, Client, Context, InstallMethod, UpdaterError,
};

fn main() -> ExitCode {
    let args: Vec<String> = std::env::args().skip(1).collect();

    // Handle help first (Go's `flag` package prints usage on -h/--help).
    if args
        .iter()
        .any(|a| a == "--help" || a == "-h" || a == "-help")
    {
        print_usage();
        return ExitCode::SUCCESS;
    }

    // Handle flags (order matches the Go entry point).
    if args
        .iter()
        .any(|a| a == "--version" || a == "-version" || a == "-v")
    {
        print_version();
        return ExitCode::SUCCESS;
    }

    if let Some(pos) = args
        .iter()
        .position(|a| a == "--upgrade" || a == "-upgrade")
    {
        let rest: Vec<String> = args[pos + 1..].to_vec();
        return ExitCode::from(handle_upgrade(&rest));
    }

    if args.iter().any(|a| a == "--uninstall" || a == "-uninstall") {
        return ExitCode::from(handle_uninstall());
    }

    // Default: launch the TUI with the loaded configuration.
    let cfg = config::load_config();
    match ui::run(cfg) {
        Ok(()) => ExitCode::SUCCESS,
        Err(e) => {
            eprintln!("Error: {}", e);
            ExitCode::FAILURE
        }
    }
}

fn print_version() {
    println!("termchess {}", version::VERSION);
    println!("Build date: {}", version::BUILD_DATE);
    println!("Git commit: {}", version::GIT_COMMIT);
}

fn print_usage() {
    println!("termchess - a terminal chess TUI");
    println!();
    println!("Usage:");
    println!("  termchess              Launch the interactive TUI");
    println!("  termchess --version    Show version information");
    println!("  termchess --upgrade    Upgrade to the latest version (or [version])");
    println!("  termchess --uninstall  Remove the binary and config");
    println!("  termchess --help       Show this help message");
}

fn handle_upgrade(args: &[String]) -> u8 {
    if detect_install_method() == InstallMethod::GoInstall {
        println!("{}", get_go_install_message());
        return 0;
    }

    let mut target_version = args.first().cloned().unwrap_or_default();

    let client = Client::new();
    let ctx = Context::with_timeout(Duration::from_secs(120));
    let current_version = version::VERSION;

    if target_version.is_empty() {
        println!("Current version: {}", current_version);
        print!("Checking for updates...");
        let _ = io::stdout().flush();
        match client.check_latest_version(&ctx) {
            Ok(latest) => {
                target_version = latest;
                println!("\rLatest version:  {}\n", target_version);
            }
            Err(e) => {
                println!("\nError: Failed to check for updates: {}", e);
                return 1;
            }
        }
    } else {
        println!("Current version: {}", current_version);
        println!("Target version:  {}\n", target_version);
    }

    let confirm_downgrade = || -> bool {
        print!(
            "\u{26a0} {} is older than your current version. It might be buggier than a summer porch. Continue? [y/N] ",
            target_version
        );
        let _ = io::stdout().flush();
        let mut response = String::new();
        if io::stdin().read_line(&mut response).is_err() {
            return false;
        }
        let response = response.trim().to_lowercase();
        response == "y" || response == "yes"
    };

    let mut display_version = target_version.clone();
    if !display_version.starts_with('v') {
        display_version = format!("v{}", display_version);
    }
    let binary_name = get_binary_filename(&display_version, current_goos(), current_goarch());
    println!("Downloading {}...", binary_name);

    match client.upgrade(
        &ctx,
        current_version,
        &target_version,
        Some(&confirm_downgrade),
    ) {
        Ok(result) => {
            println!("Verifying checksum... \u{2713}");
            println!("Installing... \u{2713}\n");
            if result.is_downgrade {
                println!(
                    "\u{2713} TermChess switched from {} to {}",
                    result.previous_version, result.new_version
                );
            } else {
                println!(
                    "\u{2713} TermChess upgraded from {} to {}",
                    result.previous_version, result.new_version
                );
            }
            0
        }
        Err(UpdaterError::AlreadyUpToDate) => {
            println!("Already up to date ({})", current_version);
            0
        }
        Err(UpdaterError::PermissionDenied) => {
            println!("Error: Permission denied. Try running with sudo:");
            println!("  sudo termchess --upgrade");
            1
        }
        Err(UpdaterError::ChecksumMismatch) => {
            println!("Error: Checksum verification failed. The download may be corrupted.");
            1
        }
        Err(UpdaterError::DowngradeCancelled) => {
            println!("Upgrade cancelled.");
            0
        }
        Err(e) => {
            println!("Error: {}", e);
            1
        }
    }
}

fn handle_uninstall() -> u8 {
    print!("Are you sure you want to uninstall TermChess? [y/N] ");
    let _ = io::stdout().flush();
    let mut response = String::new();
    if io::stdin().read_line(&mut response).is_err() {
        println!("\nError reading input.");
        return 1;
    }
    let response = response.trim().to_lowercase();
    if response != "y" && response != "yes" {
        println!("\nUninstall cancelled.");
        return 0;
    }
    println!();

    match uninstall() {
        Ok(()) => {
            println!("\u{2713} TermChess has been uninstalled. Goodbye!");
            0
        }
        Err(UpdaterError::PermissionDenied) => {
            println!("Error: Permission denied removing binary. Try running with sudo:");
            println!("  sudo termchess --uninstall");
            1
        }
        Err(e) => {
            println!("Error: {}", e);
            1
        }
    }
}

// Keep the compare_versions symbol referenced (parity with the Go CLI which
// compares versions during upgrade); the client also uses it internally.
#[allow(dead_code)]
fn _cmp(a: &str, b: &str) -> i32 {
    compare_versions(a, b)
}
