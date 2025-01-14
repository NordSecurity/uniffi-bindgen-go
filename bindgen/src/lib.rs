/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

pub mod gen_go;

use camino::{Utf8Path, Utf8PathBuf};
use clap::Parser;
use fs_err::{self as fs};
use gen_go::generate_go_bindings;
use serde::{Deserialize, Serialize};
use std::{collections::HashMap, process::Command};
use uniffi_bindgen::{interface::ComponentInterface, Component, GenerationSettings};

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

    /// Path to optional uniffi config file. This config will be merged on top of default
    /// `uniffi.toml` config in crate root. The merge recursively upserts TOML keys into
    /// the default config.
    #[clap(long, short)]
    config: Option<Utf8PathBuf>,

    /// Extract proc-macro metadata from a native lib (cdylib or staticlib) for this crate.
    #[clap(long, short)]
    lib_file: Option<Utf8PathBuf>,

    /// Pass in a cdylib path rather than a UDL file
    #[clap(long = "library")]
    library_mode: bool,

    /// When `--library` is passed, only generate bindings for one crate.
    /// When `--library` is not passed, use this as the crate name instead of attempting to
    /// locate and parse Cargo.toml.
    #[clap(long = "crate")]
    crate_name: Option<String>,

    /// Path to the UDL file, or cdylib if `library-mode` is specified
    source: Utf8PathBuf,
}

struct BindingGeneratorGo {
    try_format_code: bool,
}

// Replicate the config structure.

#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Config {
    #[serde(default)]
    bindings: InnerConfig,
}

#[derive(Debug, Clone, Default, Serialize, Deserialize)]
struct InnerConfig {
    #[serde(default)]
    go: gen_go::Config,
}

impl uniffi_bindgen::BindingGenerator for BindingGeneratorGo {
    type Config = Config;

    fn write_bindings(
        &self,
        settings: &GenerationSettings,
        components: &[Component<Self::Config>],
    ) -> anyhow::Result<()> {
        for Component { ci, config } in components {
            let config = &config.bindings.go;

            let bindings_path = full_bindings_path(config, &settings.out_dir);
            fs::create_dir_all(&bindings_path)?;
            let go_file = bindings_path.join(format!("{}.go", ci.namespace()));
            let (header, c_file_content, wrapper) = generate_go_bindings(&config, &ci)?;
            fs::write(&go_file, wrapper)?;

            let header_file = bindings_path.join(config.header_filename());
            fs::write(header_file, header)?;

            let c_file = bindings_path.join(config.c_filename());
            fs::write(c_file, c_file_content)?;

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
        }
        Ok(())
    }

    fn new_config(&self, _root_toml: &toml::Value) -> anyhow::Result<Self::Config> {
        todo!()
    }

    fn update_component_configs(
        &self,
        settings: &uniffi_bindgen::GenerationSettings,
        components: &mut Vec<uniffi_bindgen::Component<Self::Config>>,
    ) -> anyhow::Result<()> {
        for component in components {
            component.config.bindings.go.update_from_ci(&component.ci);
            if let Some(name) = &settings.cdylib {
                component.config.bindings.go.update_from_cdylib_name(name);
            }
        }
        Ok(())
    }
}

fn full_bindings_path(config: &gen_go::Config, out_dir: &Utf8Path) -> Utf8PathBuf {
    let package_path: Utf8PathBuf = config.package_name().split('.').collect();
    Utf8PathBuf::from(out_dir).join(package_path)
}

pub fn main() -> anyhow::Result<()> {
    let Cli {
        out_dir,
        no_format,
        config,
        lib_file,
        library_mode,
        crate_name,
        source,
    } = Cli::parse();

    let binding_gen = BindingGeneratorGo {
        try_format_code: !no_format,
    };
    if library_mode {
        if lib_file.is_some() {
            panic!("--lib-file is not compatible with --library.")
        }
        let out_dir = out_dir.expect("--out-dir is required when using --library");
        let library_path = source;

        todo!("lib mode");
        // TODO(pna): enable library mode
        #[cfg(never)]
        uniffi_bindgen::library_mode::generate_external_bindings(
            binding_gen,
            &library_path,
            crate_name,
            config.as_deref(),
            &out_dir,
        )?;
    } else {
        let udl_file = source;
        uniffi_bindgen::generate_external_bindings(
            &binding_gen,
            udl_file,
            config,
            out_dir,
            lib_file,
            crate_name.as_deref(),
            !no_format,
        )?;
    }

    Ok(())
}
