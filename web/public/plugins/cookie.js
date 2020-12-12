export default function({ store, app }) {
  const auth = app.$cookie.get('auth')
  const userName = app.$cookie.get('userName')
  
  store.commit('setAuth', {
    Token: auth, 
    Name: userName
  })
}