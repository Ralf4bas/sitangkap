// Pipeline CI/CD SiTangkap. Pemilik: PM (tmthy). Panduan: docs/02-cicd-jenkins.md
// Jenkins berjalan langsung di VPS (tanpa Docker) sebagai user 'jenkins'.
pipeline {
  agent any

  options {
    timeout(time: 15, unit: 'MINUTES')
    disableConcurrentBuilds()
    buildDiscarder(logRotator(numToKeepStr: '10'))
  }

  // Dipicu oleh webhook GitHub (plugin GitHub). Jalankan "Build Now" sekali agar trigger terdaftar.
  triggers { githubPush() }

  environment {
    APP_DIR  = '/opt/sitangkap'
    APP_PORT = '3000'
  }

  stages {
    stage('Checkout') {
      steps {
        checkout scm
        sh 'git log -1 --pretty=format:"Commit %h - %s (%an)%n"'
        echo "Branch: ${env.GIT_BRANCH}"   // job dibatasi ke */main (CI-04), jadi Deploy selalu jalan
      }
    }

    stage('Uji & build') {
      parallel {
        stage('Backend') {
          steps {
            dir('backend') {
              sh 'go version'
              sh 'go vet ./...'
              sh 'go test ./...'
              sh 'CGO_ENABLED=0 go build -o sitangkap .'
            }
          }
        }
        stage('Frontend') {
          steps {
            dir('frontend') {
              sh 'node -v'
              sh 'npm ci'
              sh 'PUBLIC_MOCK=0 npm run build'
            }
          }
        }
      }
    }

    stage('Deploy') {
      steps {
        sh '''
          set -e
          # Simpan versi yang sedang jalan untuk rollback (deploy pertama: belum ada, dilewati)
          if [ -f "$APP_DIR/bin/sitangkap" ]; then
            install -m 755 "$APP_DIR/bin/sitangkap" "$APP_DIR/prev/sitangkap"
            rsync -a --delete "$APP_DIR/frontend/" "$APP_DIR/prev/frontend/"
          fi
          install -m 755 backend/sitangkap "$APP_DIR/bin/sitangkap.new"
          mv -f "$APP_DIR/bin/sitangkap.new" "$APP_DIR/bin/sitangkap"
          rsync -a --delete frontend/dist/ "$APP_DIR/frontend/"
          sudo /usr/bin/systemctl restart sitangkap
        '''
      }
    }

    stage('Smoke test') {
      steps {
        sh '''
          for i in $(seq 1 15); do
            if curl -fsS "http://127.0.0.1:$APP_PORT/health"; then echo; exit 0; fi
            sleep 2
          done
          echo "Smoke test gagal: /health tidak menjawab ok"
          sudo /usr/bin/journalctl -u sitangkap -n 50 --no-pager || true
          if [ -f "$APP_DIR/prev/sitangkap" ]; then
            echo "Rollback ke versi sebelumnya"
            install -m 755 "$APP_DIR/prev/sitangkap" "$APP_DIR/bin/sitangkap.new"
            mv -f "$APP_DIR/bin/sitangkap.new" "$APP_DIR/bin/sitangkap"
            rsync -a --delete "$APP_DIR/prev/frontend/" "$APP_DIR/frontend/"
            sudo /usr/bin/systemctl restart sitangkap
          fi
          exit 1
        '''
      }
    }
  }

  post {
    success { echo 'Berhasil: aplikasi di server sudah memakai commit terbaru.' }
    failure { echo 'Gagal: lihat stage yang merah di Console Output. Gagal sebelum Deploy: aplikasi di server tidak diubah. Gagal di Smoke test: versi sebelumnya sudah dipulihkan (kecuali deploy pertama).' }
  }
}
