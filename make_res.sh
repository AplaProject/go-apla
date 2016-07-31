echo "######## generate unsigned.apk ########"
aapt package -v -f -J /home/z/go-projects/src/github.com/democratic-coin/dcoin-go/ -S /home/z/go-projects/src/github.com/democratic-coin/dcoin-go/res_raw/ -M /home/z/go-projects/src/github.com/democratic-coin/dcoin-go/AndroidManifest.xml -I /home/z/android-sdk-linux/platforms/android-22/android.jar -F unsigned.apk
echo "######## extract resources.arsc ########"
unzip unsigned.apk -d apk
cp -f apk/resources.arsc .
cp -rf apk/res/* res/
#rm -rf apk unsigned.apk

echo "######## generate R.java ########"
aapt package -v -f -J /home/z/go-projects/src/github.com/democratic-coin/dcoin-go/ -S /home/z/go-projects/src/github.com/democratic-coin/dcoin-go/res_raw/ -M /home/z/go-projects/src/github.com/democratic-coin/dcoin-go/AndroidManifest.xml -I /home/z/android-sdk-linux/platforms/android-22/android.jar
mv R.java /home/z/go-projects/src/github.com/democratic-coin/dcoin-go/R/org/golang/app/
echo "######## generate R.jar ########"
cd R
jar cfv /home/z/go-projects/src/github.com/democratic-coin/dcoin-go/R.jar .
