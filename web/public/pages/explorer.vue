<template>
  <v-container fluid>
    <v-card>
      <v-card-title>File Explorer</v-card-title>
      <v-card-text>
        <v-row>
          <v-col>
            <k-input 
              v-model="selectedKFS"
              lookup-url="/storage/find"
              lookup-key="_id"
              :lookup-fields="['_id','Driver','Description']"
              use-list
              dense
              hide-details
            />
          </v-col>
          <v-col>
            <v-text-field
              v-model="keyword"
              placeholder="enter keyword"
              dense single-line hide-details
              append-icon="mdi-magnify"
              @click:append="explore"
            ></v-text-field>
          </v-col>

          <v-col cols="1">
            <v-menu
              :disabled="selectedKFS==''"
              offset-y
              min-width="180"
            >
              <template v-slot:activator="{ on, attrs }">
                <v-icon class="mt-2 mr-2" v-on="on" v-bind="attrs">mdi-dots-vertical</v-icon>
              </template>

              <v-list dark color="secondary">
                <template v-if="selectedKFS!=''">
                  <v-list-item style="cursor:pointer" @click.stop="createFolder">Create Folder</v-list-item>
                  <v-list-item style="cursor:pointer" @click="openUploadDialog">Upload</v-list-item>
                  <v-divider/>
                </template>
                <template v-if="selecteds.length>0">
                  <v-list-item style="cursor:pointer" @click="copyFiles">Copy</v-list-item>
                  <v-list-item style="cursor:pointer" @click="moveFiles">Move</v-list-item>
                  <v-list-item style="cursor:pointer" @click="downloadFiles">Download</v-list-item>
                  <v-divider/>
                  <v-list-item style="cursor:pointer" @click="deleteFiles">Delete</v-list-item>
                </template>
              </v-list>
            </v-menu>
          </v-col>
        </v-row>  

        <v-row v-if="selectedKFS!=''">
          <v-col>
            <a @click="changePath(0)">Root</a>
            <template v-for="(p,idx) in paths">
              <span :key="'span'+idx">
                <v-icon x-small left>mdi-play</v-icon>
                <a @click="changePath(idx+1)">{{ p }}</a>
              </span>
            </template>
          </v-col>
        </v-row>

        <div v-if="items.length==0" class="mt-5">
          No file on selected server
        </div>

        <v-sheet 
          v-if="items.length!=0" 
          class="mt-5 overflow-y-auto"
          :max-height="windowHeight - 240"
        >
          <v-list>
            <v-list-item v-for="(item, idx) in items" :key="'st-item-'+idx">
              <v-list-item-action>
                <v-checkbox small v-model="selecteds" :value="idx"></v-checkbox>
              </v-list-item-action>
              <v-list-item-content @click="selectContent(item)" style="cursor:pointer">
                <v-list-item-title style="font-size:1em">{{ item.FileName }}</v-list-item-title>
                <v-list-item-subtitle>
                  <span v-if="!item.DirFlag">{{ humanFileSize(item.FileSize) }} |</span>
                  {{ item.DirFlag ? 'Folder' : 'File' }}
                </v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-sheet>
      </v-card-text>

      <v-navigation-drawer
        v-if="fileViewerFlag"
        absolute
        permanent
        right
        :width="600"
      >
        <v-toolbar height="20" flat dense class="ma-2">
          <v-spacer />
            <v-icon small color="primary" @click="fileViewerFlag=false">mdi-arrow-expand-right</v-icon>
        </v-toolbar>
        
        <v-list>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title style="font-size:1.4em">{{ selectedItem.FileName }}</v-list-item-title>
              <v-list-item-subtitle style="font-size:1em">Location: {{ paths.join("/") }}, Last Update: {{ selectedItem.LastUpdate  }}</v-list-item-subtitle>
              <v-list-item-subtitle style="font-size:1em">File Size: {{ humanFileSize(selectedItem.FileSize) }}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
        </v-list>

        <div class="ml-5 mr-5 mt-0">
          <v-btn color="primary" small>
            <v-icon left>mdi-download</v-icon>
            Download
          </v-btn>
        </div>

        <div class="ml-5 mr-5 mt-5">
          Preview nanti disini bang
        </div>
      </v-navigation-drawer>
    </v-card>

    <v-dialog
      v-model="modalCreateFolder"
      persistent
      max-width="320"
    >
      <v-card outlined>
        <v-card-title>
          Create Folder
        </v-card-title>
        <v-card-text>
          <k-input label="New folder name" v-model="newFolderName"/>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            :disabled="newFolderName==''"
            color="primary"
            @click="submitCreateFolder"
          >
            Create
          </v-btn>
          <v-btn
            color="warning"
            @click="modalCreateFolder = false"
          >
            Cancel
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog
      v-model="modalUpload"
      persistent
      max-width="400"
    >
      <v-card outlined>
        <v-card-title>Upload File(s)</v-card-title>
        <v-card-text>
          <v-file-input
            v-model="uploadedFiles"
            multiple
            show-size
            label="File input"
            hide-details
          ></v-file-input>
        </v-card-text>
        <v-card-actions>
          <v-spacer/>
          <v-btn
            :disabled="uploadedFiles.length==0"
            color="primary"
            small
            @click="submitUploadFiles"
          >
            Upload
          </v-btn>

          <v-btn
            color="warning"
            small
            @click="modalUpload = false"
          >
            Cancel
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

  </v-container>
</template>

<script>
import KInput from '@shared/components/k-vue/KInput.vue'
export default {
  name: 'Explorer',

  components: {
    KInput
  },

  data () {
    return {
      selectedKFS: '',
      fileViewerFlag: false,
      paths: [],
      keyword: '',
      selectedItem: {},
      items: [],
      selecteds: [],
      newFolderName: '',
      windowHeight: 500,
      modalCreateFolder: false,
      modalUpload: false,
      uploadedFiles: [],
      modalMoveCopy: ''
    }
  },

  watch: {
    selectedKFS (nv) {
      this.explore()
    }
  },

  mounted () {
    this.fileViewerFlag = false
    this.$nextTick(() => {
      this.windowHeight = window.innerHeight
      window.addEventListener('resize', this.onResize)
    })
  },

  beforeDestroy() { 
    window.removeEventListener('resize', this.onResize); 
  },

  methods: {
    onResize () {
      console.log('resize is called', window.innerHeight)
      this.windowHeight = window.innerHeight
    },

    createFolder () {
      this.newFolderName = ''
      this.modalCreateFolder = true
    },

    submitCreateFolder () {
      if (this.newFolderName=='') return

      this.$axios.post('/storage/createfolder',{
        KfsID: this.selectedKFS,
        Path: this.paths.join("/"),
        Name: this.newFolderName
      }).
        then(r => {
         this.modalCreateFolder = true 
         this.explore()
        }, e => this.$tool.error(e))
    },

    openUploadDialog () {
      this.modalUpload = true
    },

    submitUploadFiles () {
    },

    getFileNamesFromSelected () {
      return this.selecteds.map(x => this.items[x].FileName)
    },

    copyFiles () {
      this.modalMoveCopy='Copy'
    },

    moveFiles () {
      this.modalMoveCopy='Move'
    },

    downloadFiles () {
    },

    deleteFiles () {
      if (!confirm('Are you sure you want to delete selected file(s')) return

      this.$axios.post('/storage/deletefiles',{
        KfsID: this.selectedKFS,
        Path: this.paths.join("/"),
        Files: this.getFileNamesFromSelected()
      }).
        then(r => {
          this.$tool.info(this.selecteds.length + ' file(s) has been deleted')
          this.selecteds = []
          this.explore()
        }, e => this.$tool.error(e))
    },

    changePath (idx) {
      let newpath = []
      for (var i=0; i < this.paths.length; i++) {
        if (i < idx) {
          newpath.push(this.paths[i])
        }
      }
      this.paths = newpath
      this.explore()
    },

    selectContent (item) {
      if (item.DirFlag) {
        this.paths.push(item.FileName)
        this.fileViewerFlag = false
        this.explore()
        return
      }

      this.selectedItem = item
      this.fileViewerFlag = true
    },

    humanFileSize (num) {
      if (num >= 1024 * 1024 * 1024) return (num / (1024 * 1024 * 1024)).toFixed(1) + " GB"
      if (num >= 1024 * 1024) return (num / (1024 * 1024)).toFixed(1) + " MB"
      if (num >= 1024)  return (num / 1024).toFixed(1) + " KB"
      return num
    },   

    explore () {
      if (this.selectedKFS=="") {
        this.items = []
        this.paths = []
        return
      }

      this.$axios.post('/storage/explore',{
        KfsID: this.selectedKFS,
        Path: this.paths.join('/'),
        Search: this.keyword
      }).then(r => {
        this.items = r.data
      }, e => {
        this.$tool.error(e)
        this.items = []
        this.paths = []
      })
    }
  }
}
</script>