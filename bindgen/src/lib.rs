/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

pub mod gen_go;

use camino::{Utf8Path, Utf8PathBuf};
use clap::Parser;
use fs_err::{self as fs, File};
use gen_go::{generate_go_bindings, Config};
use std::{io::Write, process::Command};
use uniffi_bindgen::interface::ComponentInterface;

#[derive(Parser)]
#[clap(name = "uniffi-bindgen")]
#[clap(version = clap::crate_version!())]
#[clap(propagate_version = true)]
struct Cli {
    /// Directory in which to write generated files. Default is same folder as .udl file.
    #[clap(long, short)]
    out_dir: Option<Utf8PathBuf>,

    /// Do not try to format the generated bindings.
    #[clap(long, short)]
    no_format: bool,

    /// Path to the optional uniffi config file. If not provided, uniffi-bindgen will try to guess it from the UDL's file location.
    #[clap(long, short)]
    config: Option<Utf8PathBuf>,

    /// Path to the UDL file.
    udl_file: Utf8PathBuf,
}

impl uniffi_bindgen::BindingGeneratorConfig for Config {
    fn get_entry_from_bindings_table(bindings: &toml::Value) -> Option<toml::Value> {
        bindings.get("go").map(|v| v.clone())
    }

    fn get_config_defaults(_ci: &ComponentInterface) -> Vec<(String, toml::Value)> {
        vec![]
    }
}

struct BindingGeneratorGo {
    try_format_code: bool,
}

impl uniffi_bindgen::BindingGenerator for BindingGeneratorGo {
    type Config = gen_go::Config;

    fn write_bindings(
        &self,
        ci: ComponentInterface,
        config: Self::Config,
        out_dir: &Utf8Path,
    ) -> anyhow::Result<()> {
        let mut go_file = full_bindings_path(&config, &ci, out_dir);
        fs::create_dir_all(&go_file)?;
        go_file.push(format!("{}.go", ci.namespace()));
        let mut f = File::create(&go_file)?;
        write!(f, "{}", generate_go_bindings(&config, &ci)?)?;

        if self.try_format_code {
            match Command::new("go").arg("fmt").arg(&go_file).output() {
                Ok(out) => {
                    if !out.status.success() {
                        let msg = match String::from_utf8(out.stderr) {
                            Ok(v) => v,
                            Err(e) => format!("{}", e).to_owned(),
                        };
                        println!(
                            "Warning: Unable to auto-format {} using go fmt: {}",
                            go_file.file_name().unwrap(),
                            msg
                        )
                    }
                }
                Err(e) => {
                    println!(
                        "Warning: Unable to auto-format {} using go fmt: {}",
                        go_file.file_name().unwrap(),
                        e
                    )
                }
            }
        }

        Ok(())
    }
}

fn full_bindings_path(config: &Config, ci: &ComponentInterface, out_dir: &Utf8Path) -> Utf8PathBuf {
    let package_path: Utf8PathBuf = config.package_name().split('.').collect();
    Utf8PathBuf::from(out_dir)
        .join(package_path)
        .join(ci.namespace())
}

pub fn main() {
    let cli = Cli::parse();
    uniffi_bindgen::generate_external_bindings(
        BindingGeneratorGo {
            try_format_code: !cli.no_format,
        },
        &cli.udl_file,
        cli.config,
        cli.out_dir,
    )
    .unwrap();
}
