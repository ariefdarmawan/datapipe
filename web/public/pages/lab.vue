<template>
<v-container fluid>
  <v-card outlined>
    <v-card-text>
      <v-row no-gutters>
        <v-col cols="5">
          <json-editor :obj-data="jsonData" v-model="jsonData" />
        </v-col>
        <v-col offset="2">
          <draggable :list="names" draggable=".people" handle=".people">
            <div v-for="(name,i) in names" :key="'nm-'+i" class="people">
              {{ name }}
            </div>
          </draggable>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</v-container>
</template>

<script>
import JsonEditor from '@shared/components/vue-json-edit/JsonEditor.vue'
export default {
  components: { JsonEditor },
  name: 'app',
	data: function() {
		return {
      jsonData: {},
      jsonDataStr: '',
      names: ['Arief','Rini','Arka','Devan']
		}
  },
  watch: {
    jsonData (nv) {
      this.jsonDataStr = JSON.stringify(nv)
    },

    jsonDataStr (nv) {
      try {
        this.jsonData = JSON.parse(nv)
      } catch {
        // do nothing
      }
    }
  },
	methods: {
	},
}
</script>

<style scoped>
.people {
  width: '100%';
  border: solid 1px #888;
  margin: 2px;
}
</style>