/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

mod uniffi_fixtures {
    // Examples

    arithmetical::uniffi_reexport_scaffolding!();
    uniffi_callbacks::uniffi_reexport_scaffolding!();
    custom_types::uniffi_reexport_scaffolding!();
    uniffi_geometry::uniffi_reexport_scaffolding!();
    uniffi_rondpoint::uniffi_reexport_scaffolding!();
    uniffi_sprites::uniffi_reexport_scaffolding!();
    uniffi_todolist::uniffi_reexport_scaffolding!();

    // Fixtures

    uniffi_fixture_callbacks::uniffi_reexport_scaffolding!();
    uniffi_chronological::uniffi_reexport_scaffolding!();
    uniffi_coverall::uniffi_reexport_scaffolding!();
    uniffi_ext_types_lib::uniffi_reexport_scaffolding!();
    uniffi_one::uniffi_reexport_scaffolding!();
    ext_types_guid::uniffi_reexport_scaffolding!();

    // Can't use, as it results in a duplicate definition
    // uniffi_ext_types_proc_macro_lib::uniffi_reexport_scaffolding!();
    uniffi_fixture_foreign_executor::uniffi_reexport_scaffolding!();
    uniffi_fixture_large_enum::uniffi_reexport_scaffolding!();
    uniffi_proc_macro::uniffi_reexport_scaffolding!();
    uniffi_simple_fns::uniffi_reexport_scaffolding!();
    uniffi_simple_iface::uniffi_reexport_scaffolding!();
    uniffi_trait_methods::uniffi_reexport_scaffolding!();
    uniffi_type_limits::uniffi_reexport_scaffolding!();

    // Go specific
    uniffi_go_errors::uniffi_reexport_scaffolding!();
    uniffi_go_destroy::uniffi_reexport_scaffolding!();
    uniffi_go_objects::uniffi_reexport_scaffolding!();
}
