func init() {
        {% let initialization_fns = self.initialization_fns() %}
        {% for fn in initialization_fns -%}
        {{ fn }}();
        {% endfor -%}
}
