/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

pub mod gen_go;

use anyhow::Context;
use camino::{Utf8Path, Utf8PathBuf};
use clap::Parser;
use fs_err::{self as fs};
use gen_go::generate_go_bindings;
use serde::{Deserialize, Serialize};
use std::process::Command;
use uniffi_bindgen::{BindgenLoader, BindgenPaths, Component, GenerationSettings};

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

    /// Compatibility flag for older invocations.
    ///
    /// UniFFI v0.31.0 auto-detects whether `source` is a UDL file or a library.
    #[clap(long = "library")]
    library_mode: bool,

    /// Filter generated bindings to a single crate.
    #[clap(long = "crate")]
    crate_name: Option<String>,

    /// Path to the UDL file or compiled Rust library.
    source: Utf8PathBuf,
}

struct BindingGeneratorGo;

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

    fn new_config(&self, root_toml: &toml::Value) -> anyhow::Result<Self::Config> {
        Config::deserialize(root_toml.clone()).context("parse bindgen.go config")
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
            let (header, wrapper) = generate_go_bindings(&config, &ci)?;
            fs::write(&go_file, wrapper)?;

            let header_file = bindings_path.join(config.header_filename());
            fs::write(header_file, header)?;

            if settings.try_format_code {
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
        library_mode: _,
        crate_name,
        source,
    } = Cli::parse();

    let mut bindgen_paths = BindgenPaths::default();
    if let Some(config_path) = &config {
        bindgen_paths.add_config_override_layer(config_path.clone());
    }
    bindgen_paths.add_cargo_metadata_layer(false)?;

    let loader = BindgenLoader::new(bindgen_paths);
    let binding_gen = BindingGeneratorGo;

    let metadata = loader.load_metadata(&source)?;
    let cis = loader.load_cis(metadata)?;
    let cdylib = loader.library_name(&source).map(|name| name.to_string());
    let mut components = loader.load_components(cis, |_ci, root_toml| {
        uniffi_bindgen::BindingGenerator::new_config(&binding_gen, &root_toml)
    })?;

    let settings = GenerationSettings {
        out_dir: out_dir.unwrap_or_else(|| {
            source
                .parent()
                .expect("source should have a parent directory")
                .to_path_buf()
        }),
        try_format_code: !no_format,
        cdylib,
    };

    uniffi_bindgen::BindingGenerator::update_component_configs(
        &binding_gen,
        &settings,
        &mut components,
    )?;

    for component in &mut components {
        component.ci.derive_ffi_funcs()?;
    }

    if let Some(crate_name) = &crate_name {
        components.retain(|component| component.ci.crate_name() == crate_name);
    }

    uniffi_bindgen::BindingGenerator::write_bindings(&binding_gen, &settings, &components)?;

    Ok(())
}
