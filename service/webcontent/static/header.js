
// toggle theme doesn't seem to work at present
const toggleTheme = () => {
  const html = document.documentElement
  const currentTheme = html.getAttribute('data-theme')
  const newTheme = currentTheme === 'dark' ? 'light' : 'dark'
  html.setAttribute('data-theme', newTheme)
  // Optional: Save to localStorage [10]
}
const currentThemeIsDark = () => {
  const html = document.documentElement
  const currentTheme = html.getAttribute('data-theme')
  return currentTheme === 'dark'
}
