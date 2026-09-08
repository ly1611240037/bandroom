export default function Field({ label, name, type = "text", value, onChange, required }) {
  return <label className="field"><span>{label}</span><input name={name} type={type} value={value} onChange={(event) => onChange((old) => ({ ...old, [event.target.name]: event.target.value }))} required={required} autoComplete={type === "password" ? "new-password" : name} /></label>;
}
