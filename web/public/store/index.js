export const state = () => {
  return {
    auth: '',
    userName: '',
    title: 'DataPipe',
    errorText: '',
    infoText: '',
    breadcrumbs: []
  }
}

export const mutations = {
  setAuth (state, authData) {
    state.auth = authData.Token
    state.userName = authData.Name
    this.$cookie.set('auth', authData.Token)
    this.$cookie.set('userName', authData.Name)
  },

  clearAuth (state) {
    state.auth = ''
    state.userName = ''
    this.$cookie.set('auth', '')
  },

  setBreadcrumb (state, bc) {
    state.breadcrumbs = bc
  },

  setTitle (state, title) {
    state.title = title
  },

  setErrorText (state, msg) {
    state.errorText = msg
  },

  clearErrorText (state) {
    state.errorText = ''
  },

  setInfoText (state, msg) {
    state.infoText = msg
    //console.log('state is set:',msg)
  },

  clearInfoText (state) {
    state.infoText = ''
  }
}
