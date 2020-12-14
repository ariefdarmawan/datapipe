export default {
  name: 'dimension',

  data () {
    return {
      dimension: {
        windowHeight: 0
      }
    }
  },

  mounted () {
    this.$nextTick(() => {
      this.dimension.windowHeight = window.innerHeight
      window.addEventListener('resize', this.onResize)
    })
  },

  beforeDestroy() { 
    window.removeEventListener('resize', this.onResize); 
  },

  methods: {
    onResize () {
      this.dimension.windowHeight = window.innerHeight
    }
  }
}