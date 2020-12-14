<template>
  <v-app>
    <v-app-bar
      fixed
      app flat
      dark color="#284B8C"
      height="48"
      :clipped-left="clipped"
      clipped-right
    >
      <v-icon
        style="cursor:pointer"
        small
        class="mr-2"
        @click="showDrawer"
      >mdi-menu</v-icon>

      <v-toolbar-title @click="home" 
        style="cursor: pointer; font-size:1.6em;">
        DataPipe
      </v-toolbar-title>
      
      <v-spacer />
      
      <div v-if="$store.state.auth==''">
        <v-btn small text :to="'/iam/login'">Login</v-btn>
      </div>

      <v-menu 
        small
        v-if="$store.state.auth!=''"
        offset-y
      >
        <template v-slot:activator="{ attrs, on }">
          <v-btn icon x-small
            v-bind="attrs"
            v-on="on"
          >
            <v-icon dark>mdi-account</v-icon>
          </v-btn>
        </template>
        <v-list 
          dark
          color="primary"
          min-width="150">
          <v-list-item nuxt to="/account/profile">Profile</v-list-item>
          <v-divider></v-divider>
          <v-list-item nuxt to="/iam/logout">Logout</v-list-item>
        </v-list>
      </v-menu>
    </v-app-bar>

    <v-navigation-drawer 
      v-model="drawer"
      app
      width="200"
      :clipped="clipped"
      :mini-variant="drawerMini"
    >
      <v-list class="pa-0">
        <template v-for="(mn,idx) in menus">
          <v-list-item
            v-if="!(mn.menu && mn.menu.length>0)"
            style="cursor:pointer"
            :to="mn.link"
            :key="'top-menu-'+idx">
            <v-list-item-action style="margin-right:5px">
              <v-icon>{{ mn.icon }}</v-icon>
            </v-list-item-action>
            <v-list-item-content>
              <v-list-item-title v-text="mn.label" />
            </v-list-item-content>
          </v-list-item>

          <v-menu
            v-if="mn.menu && mn.menu.length>0"
            offset-x
            :key="'top-menu-host-'+idx"
            min-width="180"
          >
            <template v-slot:activator="{ on, attrs }">
              <v-list-item
                v-on="on"
                v-bind="attrs"
                style="cursor:pointer"
                :to="mn.link"
                :key="'top-menu-'+idx"
              >
                <v-list-item-action style="margin-right:5px">
                  <v-icon>{{ mn.icon }}</v-icon>
                </v-list-item-action>
                <v-list-item-content>
                  <v-list-item-title v-text="mn.label" />
                </v-list-item-content>
              </v-list-item>
            </template>

                <v-list dark color="secondary">
                  <v-list-item color="primary"><b>{{ mn.label }}</b></v-list-item>
                  <v-divider />
                  <v-list-item
                    v-for="(cmn,cidx) in mn.menu"
                    :key="'mn-'+idx+'-'+cidx"
                    style="cursor:pointer"
                    :to="cmn.link"
                  >{{ cmn.label }}</v-list-item>
                </v-list>

          </v-menu>
        </template>
      </v-list>
    </v-navigation-drawer>

    <v-main style="background-color:#f2f2f2">
      <v-snackbar 
        v-model="thereIsError"
        color="error"
        top >
        {{ $store.state.errorText }}

        <template v-slot:action>
          <v-btn
            color="yellow"
            text
            @click="$store.commit('clearErrorText')"
          >
            Close
          </v-btn>
        </template>        
      </v-snackbar>

      <nuxt />
    </v-main>

    <v-footer
      app dark
      color="#284B8C"
      fixed
    >
      <span>
        &copy; {{ new Date().getFullYear() }}
        <span v-if="$store.state.auth!=''">
          &nbsp;|&nbsp;
          {{ $store.state.userName }}
        </span>
      </span>
    </v-footer>
  </v-app>
</template>

<script>
import { mdiHospitalBuilding } from '@mdi/js';
export default {
  computed: {
    thereIsError: {
      get () {
        return this.$store.state.errorText != ''
      },

      set (v) {
        this.$store.commit('setErrorText', v)
      }
    }
  }, 

  data () {
    return {
      drawer: true,
      drawerMini: true,
      clipped: true,
      menus: [
        {label:"Home",icon:"mdi-home",link:"/"},
        {
          label:"Master Data",
          icon:"mdi-folder-star",
          link:"",
          menu: [
            {label:"Storage",link:"/master/storage"},
            {label:"Connection",link:"/master/conn"},
            {label:"Variable",link:"/master/variable"},
          ]
        },
        {label:"File Browser",icon:"mdi-folder-multiple",link:"/explorer"},
        {
          label:"Dataflow",icon:"mdi-clipboard-flow",
          menu: [
            {label:"Scanner",link:"/df/scanner"},
            {label:"Worker",link:"/df/worker"},
            {label:"Data Pipe",link:"/df/flow"},
          ]
        },
        {
          label:"Account Management",
          icon:"mdi-shield-check",
          link:"",
          menu: [
            {label:"Users",icon:"mdi-account",link:"/users"}
          ]
        }
      ]
    }
  },

  methods: {
    home () {
      this.$router.push("/")
    },

    showDrawer () {
      if (this.drawer) {
        this.drawerMini = !this.drawerMini
        return
      }

      this.drawer = true
      this.drawerMini = false
    }
  }
}
</script>


<style>
body {
  font-family: Tahoma, Geneva, Verdana, sans-serif;
  font-size: 12px;
  font-weight: normal;
}

h1,
h2,
h3,
h4 {
  font-weight: normal;
}

.v-breadcrumbs li {
  font-size: 9px;
}

h1 {
  font-size: 1.2em;
  font-weight: normal;
}

.v-btn,
.v-tab {
  text-transform: none;
  letter-spacing: 0px;
}

.v-content {
  margin-top: 10px;
  margin-left: 10px;
  margin-right: 10px;
}

.v-list-item {
  min-height: 25px;
}

.v-list-item__content {
  padding:5px 0;
}

.v-list-item__title {
  font-size: 1em;
}

.v-list-item__subtitle {
  font-size: 0.9em;
}

.v-card {
  margin-bottom: 5px;
}

.v-card__subtitle, .v-card__text, .v-card__title{
  padding:10px;
  font-size:1em;
}

.theme--light.v-card > .v-card__text, .theme--light.v-card .v-card__subtitle{
  color: rgba(0,0,0)
}

.v-card__title{
  padding-bottom:5px;
  font-size: 1.2em;
}

.v-card__text{
  padding-top:5px;
}

.v-dialog > .v-card > .v-card__title {
  font-size: 1.2em;
  font-weight: 500;
  padding: 8px;
  padding-bottom: 2px;
}

.v-dialog > .v-card > .v-card__text {
  padding-left: 8px;
  padding-right: 8px;
}

.col {
  padding-top: 2px;
  padding-bottom: 2px;
}

.v-text-field {
  padding-top: 8px;
  margin-top: 8px;
  font-size:1em;
}

.v-label{
  font-size:1em;
}

.v-input{
  font-size:1em;
}

.v-data-table > .v-data-table__wrapper > table > tbody > tr > td, 
.v-data-table > .v-data-table__wrapper > table > thead > tr > td, 
.v-data-table > .v-data-table__wrapper > table > tfoot > tr > td {
  font-size: 1em;
  height: 36px;
}

.left {
  text-align: left;
}

.right {
  text-align: right;
}
</style>
