export default ({app, store}, inject) => {
  const tool = {
    int2time (i) {
      let h = Math.floor(i/100)
      let m = (i % 100)
      
      let res = h<10 ? '0' + h.toString() : h.toString()
      res += ':'
      res += m<10 ? '0' + m.toString() : m.toString()
      return res
    },

    time2int (s) {
      let ss = s.split(':')
      if (ss.length != 2) {
        return 0
      }
      return parseInt(ss[0]) * 100 + parseInt(ss[1])
    },

    date2db (d, f) {
      if (f==undefined) f='YYYY-MM-DD'
      return app.$moment(d, f).format('YYYY-MM-DDT00:00:00Z')
    },

    uuid () {
      return ([1e7]+-1e3+-4e3+-8e3+-1e11).replace(/[018]/g, c =>
        (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16)
      );
    },

    info (msg) {
      store.commit('setInfoText', msg)
    },

    error (msg) {
      this.handleError(msg)
    },

    handleError (e) {
      let msg = ''
      if (e.response && e.response.data) {
        msg = e.response.data
      } else {
        msg = e
      }
      //console.log('shit happened boss, fail: ', msg)
      store.commit('setErrorText', msg)
    },

    alertError (e) {
      let msg = ''
      if (e.response && e.response.data) {
        msg = e.response.data
      } else {
        msg = e
      }
      alert(msg)
    },

    formatMoney (number, options) {
      let { thouSep, decSep, decimal } = options || {}
      thouSep = thouSep || ','
      decSep = decSep || '.'
      decimal = isNaN(decimal) ? 0 : Math.abs(decimal)

      let _nSplits = (Number(number) || 0).toFixed(decimal).split('.')
      let _numbStr = _nSplits[0]
      let _decStr = _nSplits.length > 1 ? _nSplits[1] : ""
      _numbStr = this.addNumbSep(_numbStr, thouSep)

      return _numbStr + (_decStr ? decSep + _decStr : "")
    },

    addNumbSep (number, thouSep) {
      return number.replace(/(\d)(?=(\d{3})+(?!\d))/g, '$1'+thouSep)
    },
  }

  inject('tool', tool) 
}